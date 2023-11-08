package librarydb

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ds-lab2-bmstu/library/core/ports/libraries"
	"ds-lab2-bmstu/pkg/readiness"
)

const probeKey = "librariesdb"

type DB struct {
	db *gorm.DB
}

func New(lg *slog.Logger, cfg libraries.Config, probe *readiness.Probe) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection to db: %w", err)
	}

	go runMigrations(lg, db, probe, cfg.MigrationInterval, cfg.EnableTestData)

	return &DB{db: db}, nil
}

func (d *DB) GetLibraries(
	_ context.Context, city string, page uint64, size uint64,
) (libraries.Libraries, error) {
	resp := libraries.Libraries{}

	tx := d.db.Begin(&sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})

	stmt := tx.Offset(int((page - 1) * size)).Limit(int(size))
	if city != "" {
		stmt = stmt.Where("city = ?", city)
	}

	var libs []Library
	if err := stmt.Find(&libs).Error; err != nil {
		tx.Rollback()

		return resp, fmt.Errorf("failed to find libraries info: %w", err)
	}

	stmt = tx.Model(&Library{})
	if city != "" {
		stmt = stmt.Where("city = ?", city)
	}

	var count int64
	if err := stmt.Count(&count).Error; err != nil {
		tx.Rollback()

		return resp, fmt.Errorf("failed to count libraries: %w", err)
	}

	resp.Total = uint64(count)
	for _, lib := range libs {
		resp.Items = append(resp.Items, libraries.Library{
			ID:      lib.ID.String(),
			Name:    lib.Name,
			Address: lib.Address,
			City:    lib.City,
		})
	}

	tx.Commit()

	return resp, nil
}

func (d *DB) GetLibrariesByIDs(
	_ context.Context, ids []string,
) (libraries.Libraries, error) {
	tx := d.db.Begin(&sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})

	var libs []Library
	if err := tx.Where("id in ?", ids).Find(&libs).Error; err != nil {
		tx.Rollback()

		return libraries.Libraries{}, fmt.Errorf("failed to find libraries info: %w", err)
	}

	resp := libraries.Libraries{Total: uint64(len(libs))}

	for _, lib := range libs {
		resp.Items = append(resp.Items, libraries.Library{
			ID:      lib.ID.String(),
			Name:    lib.Name,
			Address: lib.Address,
			City:    lib.City,
		})
	}

	tx.Commit()

	return resp, nil
}

func (d *DB) GetLibraryBooks(
	_ context.Context, libraryID string, showAll bool, page uint64, size uint64,
) (libraries.LibraryBooks, error) {
	tx := d.db.Begin(&sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})

	stmt := tx.Model(&LibraryBook{}).Where("fk_library_id = ?", libraryID)
	if !showAll {
		stmt = stmt.Where("available_count > 0")
	}

	var count int64
	if err := stmt.Count(&count).Error; err != nil {
		tx.Rollback()

		return libraries.LibraryBooks{}, fmt.Errorf("failed to get count library books info: %w", err)
	}

	stmt = tx.Offset(int((page-1)*size)).Limit(int(size)).Where("fk_library_id = ?", libraryID)
	if !showAll {
		stmt = stmt.Where("available_count > 0")
	}

	var libraryBooks []LibraryBook
	if err := stmt.Preload("BookRef").Find(&libraryBooks).Error; err != nil {
		tx.Rollback()

		return libraries.LibraryBooks{}, fmt.Errorf("failed to select library books info: %w", err)
	}

	resp := libraries.LibraryBooks{Total: uint64(count)}

	for _, book := range libraryBooks {
		resp.Items = append(resp.Items, libraries.Book{
			ID:        book.BookRef.ID.String(),
			Name:      book.BookRef.Name,
			Author:    book.BookRef.Author,
			Genre:     book.BookRef.Genre,
			Condition: book.BookRef.Condition,
			Available: book.AvailableCount,
		})
	}

	tx.Commit()

	return resp, nil
}

func (d *DB) GetLibraryBooksByIDs(
	_ context.Context, ids []string,
) (libraries.LibraryBooks, error) {
	tx := d.db.Begin(&sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})

	var books []Book

	stmt := tx.Model(&Book{}).Where("id in ?", ids)
	if err := stmt.Find(&books).Error; err != nil {
		tx.Rollback()

		return libraries.LibraryBooks{}, fmt.Errorf("failed to select library books info: %w", err)
	}

	resp := libraries.LibraryBooks{
		Total: uint64(len(books)),
	}

	for _, book := range books {
		resp.Items = append(resp.Items, libraries.Book{
			ID:        book.ID.String(),
			Name:      book.Name,
			Author:    book.Author,
			Genre:     book.Genre,
			Condition: book.Condition,
		})
	}

	tx.Commit()

	return resp, nil
}

func (d *DB) TakeBookFromLibrary(
	_ context.Context, libraryID, bookID string,
) (libraries.ReservedBook, error) {
	resp := libraries.ReservedBook{}

	tx := d.db.Begin(&sql.TxOptions{Isolation: sql.LevelSerializable})

	var libraryBook LibraryBook
	stmt := tx.Model(&libraryBook).Clauses(clause.Returning{})
	stmt = stmt.Where("fk_library_id = ?", libraryID).Where("fk_book_id = ?", bookID)

	if err := stmt.Update("available_count", gorm.Expr("available_count - 1")).Error; err != nil {
		tx.Rollback()

		// if errors.Is(err, gorm.ErrRecordNotFound) {
		// }

		return resp, fmt.Errorf("failed to update book info: %w", err)
	}

	stmt = tx.Model(&LibraryBook{}).Preload("BookRef").Preload("LibraryRef")
	if err := stmt.Where("id = ?", libraryBook.ID).First(&libraryBook).Error; err != nil {
		tx.Rollback()

		return resp, fmt.Errorf("failed to read book info: %w", err)
	}

	tx.Commit()

	resp = libraries.ReservedBook{
		Book: libraries.Book{
			ID:        libraryBook.BookRef.ID.String(),
			Name:      libraryBook.BookRef.Name,
			Author:    libraryBook.BookRef.Author,
			Genre:     libraryBook.BookRef.Genre,
			Condition: libraryBook.BookRef.Condition,
			Available: libraryBook.AvailableCount,
		},
		Library: libraries.Library{
			ID:      libraryBook.LibraryRef.ID.String(),
			Name:    libraryBook.LibraryRef.Name,
			Address: libraryBook.LibraryRef.Address,
			City:    libraryBook.LibraryRef.City,
		},
	}

	return resp, nil
}

func (d *DB) ReturnBookToLibrary(
	_ context.Context, libraryID, bookID string,
) (libraries.Book, error) {
	tx := d.db.Begin(&sql.TxOptions{Isolation: sql.LevelSerializable})

	var libraryBook LibraryBook
	stmt := tx.Model(&libraryBook).Clauses(clause.Returning{})
	stmt = stmt.Where("fk_library_id = ?", libraryID).Where("fk_book_id = ?", bookID)

	if err := stmt.Update("available_count", gorm.Expr("available_count + 1")).Error; err != nil {
		tx.Rollback()

		// if errors.Is(err, gorm.ErrRecordNotFound) {
		// }

		return libraries.Book{}, fmt.Errorf("failed to update book info: %w", err)
	}

	stmt = tx.Model(&LibraryBook{}).Preload("BookRef")
	if err := stmt.Where("id = ?", libraryBook.ID).First(&libraryBook).Error; err != nil {
		tx.Rollback()

		return libraries.Book{}, fmt.Errorf("failed to read book info: %w", err)
	}

	tx.Commit()

	resp := libraries.Book{
		ID:        libraryBook.BookRef.ID.String(),
		Name:      libraryBook.BookRef.Name,
		Author:    libraryBook.BookRef.Author,
		Genre:     libraryBook.BookRef.Genre,
		Condition: libraryBook.BookRef.Condition,
		Available: libraryBook.AvailableCount,
	}

	return resp, nil
}
