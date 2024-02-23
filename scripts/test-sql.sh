#!/bin/bash

# stop_docker name
stop_docker() {
	sudo docker stop $1 1>/dev/null 2>&1
	sudo docker rm $1 1>/dev/null 2>&1
}

# start_mysql name password port
start_mysql() {
	stop_docker $1
	sudo docker run --rm --name $1 -e MYSQL_ROOT_PASSWORD=$2 -p $3:3306 -d mysql
}

# start_postgres name password port
start_postgres() {
	stop_docker $1
	sudo docker run --rm --name $1 -e POSTGRES_PASSWORD=$2 -p $3:5432 -d postgres
}


test_sqlite() {
	rm -f test.db
	./filebrowser -a 0.0.0.0 -d sqlite3://test.db
}


test_postgres() {
	start_postgres test-postgres postgres 5433
	sleep 30
	./filebrowser -a 0.0.0.0 -d postgres://postgres:postgres@127.0.0.1:5433/postgres?sslmode=disable
}


test_mysql() {
	start_mysql test-mysql root 3307
	sleep 60
	./filebrowser -a 0.0.0.0 -d 'mysql://root:root@127.0.0.1:3307/mysql'
}

help() {
	echo "USAGE: $0 sqlite|mysql|postgres"
	exit 1
}

if (( $# == 0 )); then
	help
fi

case $1 in
	sqlite)
		test_sqlite
		;;
	mysql)
		test_mysql
		;;
	postgres)
		test_postgres
		;;
	*)
		help
esac
	

