name: Comment
fields:
  ID:
    type: AutoIncrement
  Content:
    type: String
identifiers:
  primary: ID
related:
  Commentable:
    type: ForOnePoly
    for:
      - Person
      - Company
