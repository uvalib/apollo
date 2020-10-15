# run application

# run the migrations
echo "Migrate DB"
bin/migrate -path db -verbose -database "mysql://$APOLLO_DB_USER:$APOLLO_DB_PASSWD@tcp($APOLLO_DB_HOST)/$APOLLO_DB_NAME" up
retVal=$?
if [ $retVal -ne 0 ]; then
   echo "Migration failed"
   exit $?
fi

DEVUSER_OPT=""
if [ -n "$APOLLO_DEVUSER" ]; then
   DEVUSER_OPT="-devuser $APOLLO_DEVUSER"
   echo "Set devuser to $APOLLO_DEVUSER"
fi

# run from here, since application expects web template in web/ and writes pdfs to tmp/
cd bin; ./apollo -apollo $APOLLO_HOST -dbhost $APOLLO_DB_HOST -dbname $APOLLO_DB_NAME -dbuser $APOLLO_DB_USER -dbpass $APOLLO_DB_PASSWD -iiif $APOLLO_IIIF_MAN_URL $DEVUSER_OPT

#
# end of file
#
