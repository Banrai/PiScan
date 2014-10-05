These are optional modules for automating scan-to-buy actions.

Each module looks-up a given barcode within its catalog, and produces a standard list in json matching products available for sale.

```json
[
    {
        "desc": The name or description of the product
        "sku":  The unique stock keeping unit for this vendor
        "type": One of: "UPC", "EAN", "ISBN"
        "vnd":  The vendor id
    },
	
	... 
]
```

For example, here is the result of looking up barcode <tt>619659070625</tt> in [Amazon's](amazon) U.S. catalog:

```json
[
    {
        "desc": "8gb Cruzer Fit Flash Drive Usb 3x5 Inches Non-Retail Packaging", 
        "sku": "B00FL5UTLY", 
        "type": "UPC", 
        "vnd": "AMZN:us"
    }, 
    {
        "desc": "SanDisk Cruzer Fit 8GB USB 2.0 Low-Profile Flash Drive- SDCZ33-008G-B35", 
        "sku": "B005FYNSUA", 
        "type": "UPC", 
        "vnd": "AMZN:us"
    }, 
    {
        "desc": "SanDisk (SDCZ33-008G-B35) 6-Pack 8GB Cruzer Fit USB 2.0 Flash Drive - Encryption Support, Password Protection", 
        "sku": "B00KBLBOYY", 
        "type": "UPC", 
        "vnd": "AMZN:us"
    }
]
``` 
