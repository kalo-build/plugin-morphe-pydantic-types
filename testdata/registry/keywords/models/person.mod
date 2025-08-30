name: Person
fields:
  ID:
    type: AutoIncrement
  class:
    type: String
  for:
    type: String
  pass:
    type: String
  import:
    type: String
identifiers:
  primary: ID
related:
  from:
    type: HasOne
    target: Company
