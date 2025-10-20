-- Databricks SQL DDL demonstrating complex nested types
-- This schema showcases STRUCT and ARRAY types with various nesting patterns

CREATE OR REPLACE TABLE orders (
  order_id STRING COMMENT 'Unique identifier for each order',
  order_date TIMESTAMP COMMENT 'Date and time when the order was placed',
  total_amount DECIMAL(10, 2) COMMENT 'Total order amount in the base currency',
  
  customer STRUCT<
    customer_id: STRING COMMENT 'Unique customer identifier',
    email: STRING COMMENT 'Customer email address',
    loyalty_tier: STRING COMMENT 'Customer loyalty program tier (Bronze, Silver, Gold, Platinum)'
  > COMMENT 'Customer information as a simple struct',
  
  shipping_address STRUCT<
    street: STRING COMMENT 'Street address',
    city: STRING COMMENT 'City name',
    state: STRING COMMENT 'State or province',
    postal_code: STRING COMMENT 'Postal or ZIP code',
    country: STRING COMMENT 'Country name',
    coordinates STRUCT<
      latitude: DOUBLE COMMENT 'Latitude coordinate',
      longitude: DOUBLE COMMENT 'Longitude coordinate'
    > COMMENT 'Geographic coordinates for delivery location'
  > COMMENT 'Shipping address with nested coordinate struct',
  
  applied_coupon_codes ARRAY<STRING> COMMENT 'List of coupon codes applied to this order',
  
  line_items ARRAY<STRUCT<
    item_id: STRING COMMENT 'Unique item identifier',
    product_name: STRING COMMENT 'Name of the product',
    quantity: INT COMMENT 'Number of units ordered',
    unit_price: DECIMAL(10, 2) COMMENT 'Price per unit',
    discount_percentage: DOUBLE COMMENT 'Discount percentage applied to this item'
  >> COMMENT 'Order line items with product details and pricing',
  
  payment_info STRUCT<
    payment_method: STRING COMMENT 'Payment method used (credit card, PayPal, etc.)',
    currency: STRING COMMENT 'Currency code (USD, EUR, GBP, etc.)',
    billing_address STRUCT<
      street: STRING COMMENT 'Billing street address',
      city: STRING COMMENT 'Billing city',
      postal_code: STRING COMMENT 'Billing postal code'
    > COMMENT 'Billing address for payment',
    transactions ARRAY<STRUCT<
      transaction_id: STRING COMMENT 'Unique transaction identifier',
      amount: DECIMAL(10, 2) COMMENT 'Transaction amount',
      status: STRING COMMENT 'Transaction status (pending, completed, failed, refunded)',
      timestamp: TIMESTAMP COMMENT 'Transaction timestamp'
    >> COMMENT 'List of payment transactions for this order'
  > COMMENT 'Payment information including method, billing, and transaction history',
  
  fulfillment_events ARRAY<STRUCT<
    event_type: STRING COMMENT 'Type of fulfillment event (picked, packed, shipped, delivered)',
    event_timestamp: TIMESTAMP COMMENT 'When the event occurred',
    warehouse_location: STRING COMMENT 'Warehouse or distribution center code',
    tracking_numbers: ARRAY<STRING> COMMENT 'Shipping carrier tracking numbers',
    packages ARRAY<STRUCT<
      package_id: STRING COMMENT 'Unique package identifier',
      weight_kg: DOUBLE COMMENT 'Package weight in kilograms',
      dimensions_cm STRUCT<
        length: INT COMMENT 'Package length in centimeters',
        width: INT COMMENT 'Package width in centimeters',
        height: INT COMMENT 'Package height in centimeters'
      > COMMENT 'Package dimensions for shipping calculations'
    >> COMMENT 'List of packages in this fulfillment event'
  >> COMMENT 'Order fulfillment history with packaging and shipping details'
)
USING DELTA
COMMENT 'E-commerce orders with complex nested data structures';
