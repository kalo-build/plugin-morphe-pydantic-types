name: Contact
fields:
  ID:
    type: AutoIncrement
  Email:
    type: String
  Phone:
    type: String
identifiers:
  primary: ID
related:
  Person:
    type: ForOne
    target: Person
