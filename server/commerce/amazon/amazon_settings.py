#!/usr/bin/env python

"""
amazon_settings.py

This file contains the Amazon Product API credentials necessary for
access.

It imports everything it needs from a private, local file
(amazon_local_settings.py) which should not be checked into source
control.

"""

ACCESS_KEY = ''
SECRET_KEY = ''
ASSOCIATE  = '' 
AMZLOCALE  = 'us'

try:
    from amazon_local_settings import *
except ImportError:
    pass
