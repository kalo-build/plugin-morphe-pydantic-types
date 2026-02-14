name: Company
fields:
  ID:
    type: AutoIncrement
  Name:
    type: String
  TaxID:
    type: String
identifiers:
  primary: ID
  name: Name
related:
  Person:
    type: HasMany