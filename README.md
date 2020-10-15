# Apollo

Apollo is a management system for digitized collections.

### System Requirements
* GO version 1.12 or greater (mod required)
* Node version 8.11.1 or higher (https://nodejs.org/en/)
* Yarn version 1.10 or greater
* vue-cli 3 version 3.0.5 or greater
* Vue 2.5 or greater
* MySQL 5.0 or greater

### Build Instructions

1. After clone, `cd frontend` and execute `yarn install` to prepare the front end
2. Create a new MySQL database/user to hold the apollo data tables (apollo/apollo is expected)
3. Initilize the schema: `mysql apollo < db/v1.sql` (assuming the instance is named apollo)
4. Run the default Makefile target to build binaries for linux, darwin and the front end.  All results will be in the bin directory.

There are two commands built; the server itself (apollosvr) and a data ingest utility (apolloingest). Both require several environment variables to run:

* `APOLLO_DB_HOST` - the host where MySQL is running (including port)
* `APOLLO_DB_NAME` - the name of the apollo db instance (usually appollo)
* `APPOLO_DB_USER` - MySQL user with full permission on the DB
* `APOLLO_DB_PASS` - MySQL user password

All of these env variables can be passed as command-line args too. The are - dbhost, dbname, dbuser and dbpass.

Before running the server, run apolloingest with one or more of the data files from db/data to provide some starting data.
For example: `./bin/apolloingest.darwin -src=db/data/mountainwork.xml`

### Current API

* GET /version : return service version info
* GET /healthcheck : test health of system components; results returned as json
* GET /api/search : Search for the term provided in the query string
* GET /api/types : Get a json list of registered node types
* GET /api/values/:type : Get a json list of controlled values for a given node type
* GET /api/collections : get a json list of collections
* GET /api/collections/:PID : Get full details for the specified collection as json
* GET /api/aries : Aries ping request
* GET /api/aries/:ID : return apollo info for the specified ID

### Notes

To run in a develoment mode for the frontend only, build and launch apollosvr, and suplly a non-defalt port param: `./apollosvr.linux -port 8085`
Change directory to the frontend and add another ENV variable:

* `APOLLO_API` - set to localhost at the port from above.

Other environment variables are set in ./frontent/.env (production) and ./frontend/.env.development (dev overrides)

Run the frontend in development mode with `yarn serve` from the frontend directory. All API resests from the frontend will be rediected to the local instance of the backend services.

To run the backend in dev mode (without Shibboleth) you will also need to supply a fake devuser identifer with the launch command: `./apollosvr.linux -port 8085 -devuser x5ht`

Run migrations like this:

`migrate -database ${APOLLO_DB} -path backend/db/migrations up`

Example migrate commads to create a migration and run one:

* `migrate create -ext sql -dir backend/db/migrations -seq update_user_auth`
* `migrate -database ${APOLLO_DB} -path backend/db/migrations/ up`
