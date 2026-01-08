# Story-001: 基础设施搭建

> 创建日期: 2026-01-07
> 负责人: cooperswang
> 状态: 🔄 未开始

## 🎯 目标

搭建 OIP Backend 的 swagger 文档，并验证API 文档生成是否正确。

## 📋 任务拆解

- [ ] 1. Plan 模式对齐（Claude Session 1）
- [ ] 2. 参考接口文档（见 参考 API 文档）
- [ ] 3. 阅读项目相关文档并整理 Schema 框架
  - [ ] `accounts` 
  - [ ] `orders` 
- [ ] 4. 编写 swagger 功能文档
  - [ ] Account 接口文档
  - [ ] orders 接口文档
- [ ] 5. 本地测试验证

## ✅ 验证标准

- [ ] 本地启动文档验证，检查各个字段定义是否清晰完整

## 🤖 Claude 会话记录

## 📝 开发笔记

- `/Users/cooperswang/Documents/wsqyouth/oip/backend`项目路径
- 

## ⚠️ 遇到的问题与解决方案

### 问题 1: [待记录]
- **现象**: -
- **原因**: -
- **解决方案**: -

## 📦 交付物

- [ ] Swagger 文档

## 🔗 相关链接

## 📊 性能指标

## 🎓 经验总结

[待完成后总结]



### 参考 API 文档

* 预期效果

  ![image-20260107164735194](/Users/cooperswang/Library/Application Support/typora-user-images/image-20260107164735194.png)

```
Create a label
post
https://sandbox-api.aftership.com/postmen/v3
/labels
Create a label.

Request
An API key is a token that you provide when making API calls. Include the token in a header parameter called as-api-key.

Example: as-api-key: 123

Headers
Content-Type
string
required
Content-Type

Default:
application/json
Example:
application/json
Body
application/jsonapplication/xml

application/json
Create a label object.

billing
Billing
Billing object: the description of billing information

paid_by
string
required
Allowed values:
shipper
third_party
recipient
method
PaymentMethodAccount
PaymentMethodAccount object: the description of account information

customs
Customs
Customs object: the description of customs information

purpose
string
required
Allowed values:
gift
merchandise
personal
sample
return
repair
non-merchandise
terms_of_trade
string
Allowed values:
dat
ddu
ddp
dap
exw
fca
fob
cip
cif
cpt
cfr
dpu
eei
AESNOEEI

one of: AES
AES object: the description of EEI Type - aes

billing
Billing
Billing object: the description of billing information

importer_address
Address
Address object: the description of address information

broker_address
Address
Address object: the description of address information

passport
object
Passport Object

additional_charges
array[object]
This array contains additional_charges object. Additional charge to be added to the commercial invoice of this shipment. Only applies to FedEx, DHL Express, UPS currently.

<= 3 items
return_shipment
boolean
(set to false if not provided)

is_document
boolean
If the shipment is document type. (set to false if not provided)

service_type
string
paper_size
string
shipper_account
object
id
string
references
array[string]
shipment
Shipment
Shipment object: the description of shipment information

tax_total
Money
The money object is used to represent the currency amount.

ship_from
Address
required
Address object: the description of address information

ship_from_display
Address
Address object: An optional address object that lets you customize the ship from details shown on the shipping label.
This is only for visual display and won't change actual shipping routes or carrier operations.
Supported slug: purolator,canada-post

ship_to
Address
required
Address object: the description of address information

street1
string
required
address_line1 of address

country
string
required
Country in ISO 3166-1 alpha 3 code

contact_name
string
required
contact_name of address

phone
string
contact_phone of address

fax
string or null
fax_number of address

email
string or null
email of address

company_name
string or null
company_name of address

street2
string or null
address_line2 of address

street3
string or null
address_line3 of address

city
string or null
city of address

state
string or null
state of address

postal_code
string or null
postal_code of address

type
string or null
type of address

Allowed values:
residential
business
tax_id
string or null
tax id

tax_id_type
string or null
tax id type. Only applies to DHL Express, DHL eCommerce US, DHL eCommerce UK.

vat: Value-Added tax
gst: Goods and Service Tax
eori: European Union Registration and Identification
ioss: Import One Stop Shop
pan: Pan-Aarav
ukims: UK Internal Market Scheme
other: Other
Allowed values:
vat
gst
eori
ioss
pan
ukims
other
identification_number
string or null
identification number

identification_type
string or null
identification type

mil: Military Number
nid: National Identity Card
pas: Passport
other: Other
Allowed values:
mil
nid
pas
other
eori_number
string or null
eori number

location
object or null
parcels
array[Parcel]
required
Parcels of shipment

box_type
string
required
Type of box for packaging

dimension
Dimension
required
Dimension object: the description of width/height/depth information

items
array[Item]
required
items of package Item object, use to describe product to ship

description
string
The description of the parcel

weight
Weight
Weight object: unit weight of the item

return_to
Address
Address object: the description of address information

delivery_instructions
string
Instructions to aid in prompt delivery of the shipment.

invoice
Invoice or null
Invoice object: the description of invoice type, commercial invoice will be generated if field is present in the request body

date
string
required
Invoice date in YYYY-MM-DD

number
string
Invoice number

type
string
Allowed values:
commercial
proforma
number_of_copies
integer
>= 1
<= 4
signature_name
string
Signature name showing on invoice. Only applies to FedEx, DHL Express currently.

declaration_statement
string
Declaration statement showing on invoice. Only applies to FedEx, DHL Express currently.

service_options
array[object]
This array contains service_options object. Please refer to all service options and service types and service options list for details.

file_type
string
Allowed values:
pdf
zpl
ship_date
string
Ship Date in YYYY-MM-DD, if not provided, will be today of the shipper account timezone

order_number
string
A user-defined order number used to identify an order.

order_id
string
Unique identification of the order.

custom_fields
object
Custom fields that accept an object with string field.Show all...

files
InputFiles
Files object: additional shipping documents to upload with label generation. Only applies to FedEx, DHL Express currently.

certificate_of_origin
object
Certificate of Origin

commercial_invoice
object
Commercial Invoice

customs_declaration
object
Customs Declaration

print_options
object
qr_code
object
Whether to return the qr_code when creating a label

Supported slug: arvato, canada-post, dpd-uk, evri, vesyl, usps, poste-italiane, pitney-bowes, dpd-nl
Responses
200
OK

Body

application/json

application/json
responses
/
200
meta
Meta
Meta data object.

Example:
{"code":200,"type":"OK","message":"Everything worked as expected."}
code
integer
required
Code of Meta

message
string
required
Message of Meta

details
array[Error]
Details of Meta

retryable
boolean
Whether this request will be retryable or not

data
Label
Label object: the description of label object

id
string
Label ID

status
string
Allowed values:
creating
created
cancelling
cancelled
manifesting
manifested
failed
ship_date
string
tracking_numbers
array[string]
carrier_references
array[object]
files
Files
Files object: the description of 6 different file objects(label, qr_code, invoice, customs_declaration, packing_slip, manifest)

rate
Rate
Rate object: the description of rate information

created_at
string
A formatted date.

updated_at
string
A formatted date.

references
array[string] or null
Reference information for a label

shipper_account
Reference
Reference object: the description of reference information

service_type
string
Service Types

order_number
string
A user-defined order number used to identify an order.

service_options
array[object]
This array contains service_options object. Please refer to all service options and service types and service options list for details.

order_id
string
Unique identification of the order.

custom_fields
object
Custom fields that accept an object with string field.Show all...

carrier_redirect_link
string or null
Delivery instructions (delivery date or address) can be modified by visiting the link if supported by a carrier.
```

* 每个接口示意效果
  ![image-20260107164813033](/Users/cooperswang/Library/Application Support/typora-user-images/image-20260107164813033.png)

```
Create a shipper account
post
https://sandbox-api.aftership.com/postmen/v3
/shipper-accounts
This endpoint allows you to create your shipper account via API. You can integrate our API with your system and manage the shipper account according to your needs.

Please ensure that your account credentials are correct. Different carriers have different requirements for account credentials, you can refer to Shipper account's credentials for more details.

Request
An API key is a token that you provide when making API calls. Include the token in a header parameter called as-api-key.

Example: as-api-key: 123

Headers
Content-Type
string
required
Content-Type

Default:
application/json
Example:
application/json
Body
application/jsonapplication/xml

application/json
Create a shipper account

slug
string
description
string
timezone
string
credentials
object
Please refer to Shipper account's credentials

address
Address
Address object: the description of address information

street1
string
required
address_line1 of address

country
string
required
Country in ISO 3166-1 alpha 3 code

contact_name
string
required
contact_name of address

phone
string
contact_phone of address

fax
string or null
fax_number of address

email
string or null
email of address

company_name
string or null
company_name of address

street2
string or null
address_line2 of address

street3
string or null
address_line3 of address

city
string or null
city of address

state
string or null
state of address

postal_code
string or null
postal_code of address

type
string or null
type of address

Allowed values:
residential
business
tax_id
string or null
tax id

tax_id_type
string or null
tax id type. Only applies to DHL Express, DHL eCommerce US, DHL eCommerce UK.

vat: Value-Added tax
gst: Goods and Service Tax
eori: European Union Registration and Identification
ioss: Import One Stop Shop
pan: Pan-Aarav
ukims: UK Internal Market Scheme
other: Other
Allowed values:
vat
gst
eori
ioss
pan
ukims
other
identification_number
string or null
identification number

identification_type
string or null
identification type

mil: Military Number
nid: National Identity Card
pas: Passport
other: Other
Allowed values:
mil
nid
pas
other
eori_number
string or null
eori number

location
object or null
settings
object
Settings object: additional settings of the shipper account. Only applies to FedEx currently.

commercial_invoice_letterhead
string
Letterhead image to generate courier's electronic commercial invoice. The max resolution is 700 pixels wide by 50 pixels long.

commercial_invoice_signature
string
Signature image to generate courier's electronic commercial invoice. The max resolution is 700 pixels wide by 50 pixels long.

Responses
200
Create a shipper account

Body

application/json

application/json
responses
/
200
meta
Meta
Meta data object.

Example:
{"code":200,"type":"OK","message":"Everything worked as expected."}
code
integer
required
Code of Meta

message
string
required
Message of Meta

details
array[Error]
Details of Meta

retryable
boolean
Whether this request will be retryable or not

data
ShipperAccount
ShipperAccount object: the description of shipper account information

id
string
Shipper Account ID

address
Address
Address object: the description of address information

slug
string
Couirer slug

status
string
Allowed values:
enabled
disabled
deleted
description
string
The description of the account

type
string
The type of the shipper_account

Allowed value:
default
timezone
string
Shipper account timezone

settings
object
Settings object: additional settings of the shipper account. Only applies to FedEx currently.

created_at
string
A formatted date.

updated_at
string
A formatted date.
```

```
curl --request POST \
  --url https://sandbox-api.aftership.com/postmen/v3/shipper-accounts \
  --header 'Content-Type: application/json' \
  --header 'as-api-key: ' \
  --data '{
  "slug": "dhl",
  "description": "My Shipper Account",
  "timezone": "Asia/Hong_Kong",
  "credentials": {
    "account_number": "******",
    "password": "******",
    "site_id": "******"
  },
  "address": {
    "contact_name": "AfterShip Shipping",
    "company_name": "AfterShip Shipping",
    "street1": "230 W 200 S LBBY",
    "city": "Salt Lake City",
    "state": "UT",
    "postal_code": "84101",
    "country": "USA",
    "phone": "123456789",
    "email": "test@test.com"
  }
}'
```



