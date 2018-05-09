package GoWars

import (
	"log"
)

var (
	DB ItemDatabase
)

func init() {
	var err error

	// [START cloudsql]
	DB, err = configureCloudSQL(cloudSQLConfig{
		Username: "root",
		Password: "burg1996",
		Instance: "gsc-gowars:europe-west1:gowars",
	})
	// [END cloudsql]

	if err != nil {
		log.Fatal(err)
	}
}

type cloudSQLConfig struct {
	Username, Password, Instance string
}

func configureCloudSQL(config cloudSQLConfig) (ItemDatabase, error) {
	//if os.Getenv("GAE_INSTANCE") != "" {
	//	// Running in production.
	//	return newMySQLDB(MySQLConfig{
	//		Username:   config.Username,
	//		Password:   config.Password,
	//		UnixSocket: "/cloudsql/" + config.Instance,
	//	})
	//}

	// Running locally.
	return newMySQLDB(MySQLConfig{
		Username: config.Username,
		Password: config.Password,
		Host:     "35.205.161.100",
		Port:     3306,
	})
}
