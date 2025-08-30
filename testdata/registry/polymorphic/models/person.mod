name: Person
fields:
  ID:
    type: AutoIncrement
  Name:
    type: String
identifiers:
  primary: ID
related:
  Comments:
    type: HasManyPoly
    through: Commentable
  ContactInfo:
    type: HasOne
    target: Contact
    aliased: Contact
