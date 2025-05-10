package main

import (
	"log"
	"testing"
)

func TestMain(m *testing.M) {
	IsTestMode = true
	if err := InitTestDB(); err != nil {
		log.Fatal("ошибка подключения к БД: ", err)
	}
	defer TestGuestDB.Close()
	defer TestAdminDB.Close()
	defer TestUserDB.Close()

	if err := CreateAll(TestAdminDB); err != nil {
		log.Fatal("ошибка создания таблиц БД: ", err)
	}

	m.Run()

	// if err := DeleteAll(TestAdminDB); err != nil {
	// 	log.Fatal("ошибка удаления таблиц БД: ", err)
	// }
}
