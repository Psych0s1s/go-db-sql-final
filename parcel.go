package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {

	statement := `INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :created_at)`

	stmt, err := s.db.Prepare(statement)
	if err != nil {
		return 0, fmt.Errorf("error preparing insert statement: %v", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(sql.Named("client", p.Client), sql.Named("status", p.Status), sql.Named("address", p.Address), sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, fmt.Errorf("error executing insert statement: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert ID: %v", err)
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {

	statement := `SELECT number, client, status, address, created_at FROM parcel WHERE number = :number`

	stmt, err := s.db.Prepare(statement)
	if err != nil {
		return Parcel{}, fmt.Errorf("error preparing select statement: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(sql.Named("number", number))

	var p Parcel

	err = row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Parcel{}, fmt.Errorf("no parcel found with number %d", number)
		}
		return Parcel{}, fmt.Errorf("error scanning parcel: %v", err)
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {

	statement := `SELECT number, client, status, address, created_at FROM parcel WHERE client = :client`

	stmt, err := s.db.Prepare(statement)
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(sql.Named("client", client))
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	var parcels []Parcel

	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		parcels = append(parcels, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %v", err)
	}

	return parcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {

	validStatuses := map[string]bool{
		ParcelStatusRegistered: true,
		ParcelStatusSent:       true,
		ParcelStatusDelivered:  true,
	}

	if _, ok := validStatuses[status]; !ok {
		return fmt.Errorf("invalid status '%s' provided", status)
	}

	statement := `UPDATE parcel SET status = :status WHERE number = :number`

	stmt, err := s.db.Prepare(statement)
	if err != nil {
		return fmt.Errorf("error preparing update statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(sql.Named("status", status), sql.Named("number", number))
	if err != nil {
		return fmt.Errorf("error executing update statement with named parameters: %v", err)
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {

	statement := `UPDATE parcel SET address = :address WHERE number = :number AND status = :status`

	stmt, err := s.db.Prepare(statement)
	if err != nil {
		return fmt.Errorf("error preparing update statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(sql.Named("address", address), sql.Named("number", number), sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		return fmt.Errorf("error executing update statement with named parameters: %v", err)
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {

	statement := `DELETE FROM parcel WHERE number = :number AND status = :status`

	stmt, err := s.db.Prepare(statement)
	if err != nil {
		return fmt.Errorf("error preparing delete statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(sql.Named("number", number), sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		return fmt.Errorf("error executing delete statement with named parameters: %v", err)
	}

	return nil
}
