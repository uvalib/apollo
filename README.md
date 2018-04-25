# Apollo

Apollo is a management system for digitized collections.

### System Requirements
* GO version 1.9.2 or greater
* DEP (https://golang.github.io/dep/) version 0.4.1 or greater
* NPM version 6.0.0 or higher ( https://www.npmjs.com/get-npm )
* Node version 8.11.1 or higher (https://nodejs.org/en/)
* Vue 2.0 or greater
* MySQL 5.0 or greater

### Build Instructions

1. After clone, `cd frontend` and execute `npm install` to prepare the front end
2. Create a new MySQL database/user to hold the apollo data tables (apollo/apollo is expected)
3. Initilize the schema: `mysql apollo < db/v1.sql` (assuming the instance is named apollo)
4. Run the default Makefile target to build binaries for linux, darwin and the front end.  All results will be in the bin directory.

There are two commands built; the server itself (apollosvr) and a data ingest utility (apolloingest). Both require several environment variables to run:

Params for both:

* `APOLLO_DB_HOST` - the host where MySQL is running (including port)
* `APOLLO_DB_NAME` - the name of the apollo db instance (usually appollo)
* `APPOLO_DB_USER` - MySQL user with full permission on the DB 
* `APOLLO_DB_PASS` - MySQL user password

Params for service only:

* `APOLLO_PORT` - The port at which to run the service (optional; default 8080)
* `APOLLO_HTTPS` -  0 to serve over http, 1 to serve https 
* `APOLLO_KEY` - SSH Key to use if HTTPS is enabled
* `APOLLO_CRT` - SSH Cxrt to use if HTTPS is enabled

Before running the server, run apolloingest with one or more of the data files from db/data to provide some starting data. 
For example: `./bin/apolloingest.darwin -src=db/data/mountainwork.xml`

### Current API

* GET /version : return service version info
* GET /healthcheck : test health of system components; results returned as json
* GET /api/users : Get a json list of system users 
* GET /api/users/:ID : Get json details for a user
* GET /api/collections : get a json list of collections 
* GET /api/collections/:PID : Get full details for the specified collection as json 

### Notes

To run in a develoment mode for the frontend only, build and launch apollosvr, and suplly a non-defalt port param: `./apollosvr.linux -port 8085` 
Change directory to the frontend and add another ENV variable:

* `APOLLO_API` - set to localhost at the port from above.

Run the frontend in development mode with `npm run dev` from the frontend directory. All API resests from the frontend will be rediected to the local instance of the backend services.

