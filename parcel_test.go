package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// get
	retrievedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, retrievedParcel.Client)
	require.Equal(t, parcel.Status, retrievedParcel.Status)
	require.Equal(t, parcel.Address, retrievedParcel.Address)
	require.Equal(t, parcel.CreatedAt, retrievedParcel.CreatedAt)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)

	// verify delete
	_, err = store.Get(id)
	require.Error(t, err) // Ожидаем ошибку, так как посылка должна быть удалена
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	updatedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, updatedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// set status
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// check
	updatedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, updatedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := make(map[int]Parcel)

	clientID := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = clientID
	}

	// add
	for i := range parcels {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id, "Parcel ID should not be zero")
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(clientID)
	require.NoError(t, err, "Fetching parcels by client failed")
	require.Len(t, storedParcels, len(parcels), "The number of retrieved parcels does not match the expected count")

	// check
	for _, retrievedParcel := range storedParcels {
		expectedParcel, found := parcelMap[retrievedParcel.Number]
		require.True(t, found, "Retrieved parcel was not found in the expected map")
		require.Equal(t, expectedParcel.Client, retrievedParcel.Client, "Client IDs do not match")
		require.Equal(t, expectedParcel.Address, retrievedParcel.Address, "Addresses do not match")
		require.Equal(t, expectedParcel.Status, retrievedParcel.Status, "Statuses do not match")
		require.Equal(t, expectedParcel.CreatedAt, retrievedParcel.CreatedAt, "Creation timestamps do not match")
	}
}
