# run application

DEVUSER_OPT=""
if [ -n "$APOLLO_DEVUSER" ]; then
   DEVUSER_OPT="-devuser $APOLLO_DEVUSER"
   echo "Set devuser to $APOLLO_DEVUSER"
fi

# run from here, since application expects web template in web/ and writes pdfs to tmp/
cd bin; ./apollo -host $APOLLO_HOST -dbhost $APOLLO_DB_HOST -dbname $APOLLO_DB_NAME -dbuser $APOLLO_DB_USER -dbpass $APOLLO_DB_PASSWD -iiif $APOLLO_IIIF_MAN_URL -solr_dir $APOLLO_SOLR_DROPOFF_DIR -qdc_dir $APOLLO_QDC_DELIVERY_DIR -fedora $APOLLO_WSLS_FEDORA_URL $DEVUSER_OPT

#
# end of file
#