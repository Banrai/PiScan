#!/usr/bin/env python

"""
amazon_api_lookup.py

This module uses the Amazon Product API to lookup barcodes as UPC and
EAN codes, and returns a list of possible product matches in the Amazon
catalog, along with the corresponding Amazon Standard Identification
Number (ASIN).

The actual Amazon Product API credentials necessary for access are
stored in a private, local file (amazon_local_settings.py) via a settings
harness.

"""

import sys
import json

from amazonproduct import API, errors

from amazon_settings import ACCESS_KEY, SECRET_KEY, ASSOCIATE, AMZLOCALE


api = API(locale=AMZLOCALE, access_key_id=ACCESS_KEY, secret_access_key=SECRET_KEY, associate_tag=ASSOCIATE)

def lookup (barcode, ID_TYPES=['UPC','EAN', 'ISBN']):
    """Lookup the given barcode and return a list of possible matches"""

    matches = [] # list of {'desc', 'asin', 'type'}

    for idtype in ID_TYPES:
        try:
            result = api.item_lookup(barcode, SearchIndex='All', IdType=idtype)
            for item in result.Items.Item:
                matches.append({'desc': unicode(item.ItemAttributes.Title),
                                'asin': unicode(item.ASIN),
                                'type': idtype})

        except (errors.InvalidAccount, errors.InvalidClientTokenId, errors.MissingClientTokenId):
            print >>sys.stderr, "Amazon Product API lookup: bad account credentials"

        except errors.TooManyRequests, toomanyerr:
            print >>sys.stderr, "Amazon Product API lookup error:", toomanyerr

        except errors.InternalError, awserr:
            print >>sys.stderr, "Amazon Product API lookup error:", awserr

        except errors.InvalidParameterValue:
            # this simply means the barcode
            # does not exist for the given type,
            # so no need to do anything explicit
            pass

    return matches


if __name__ == "__main__":
    """Create a command-line main() entry point"""

    if len(sys.argv) != 2:
        # Define the usage 
        print sys.argv[0], '[barcode]'

    else:
        # lookup the barcode and return
        # a string of the results to stdout,
        # or nothing if there were no matches

        products = lookup(sys.argv[1])
        if len(products) > 0:
            print json.dumps(products)

