package main

import (
	"github.com/krijnrien/GoWars/gowars_db"
	"log"
)

var DB *gowars_db.MySQLConn

type cloudSQLConfig struct {
	Username, Password, Instance string
}

func init() {
	var configErr error

	DB, configErr = configureCloudSQL()
	if configErr != nil {
		log.Fatalln(configErr)
	}
}

func configureCloudSQL() (*gowars_db.MySQLConn, error) {
	//if os.Getenv("GAE_INSTANCE") != "" {
	//	// Running in production.
	//	return newMySQLDB(MySQLConfig{
	//		Username:   config.Username,
	//		Password:   config.Password,
	//		UnixSocket: "/cloudsql/" + config.Instance,
	//	})
	//}

	//

	//config := cloudSQLConfig{
	//	Username: "root",
	//	Password: "burg1996",
	//	//Instance: "gsc-gowars:europe-west1:gowars",
	//}
	//
	//// Running locally.
	//conf := gowars_db.MySQLConfig{
	//	Username: config.Username,
	//	Password: config.Password,
	//	Host:     "35.205.161.100",
	//	Port:     3306,
	//}

	config := cloudSQLConfig{
		Username: "root",
		Password: "",
		//Instance: "gsc-gowars:europe-west1:gowars",
	}

	// Running locally.
	conf := gowars_db.MySQLConfig{
		Username: config.Username,
		Password: config.Password,
		Host:     "localhost",
		Port:     3306,
	}

	return conf.NewMySQLDB()
}
