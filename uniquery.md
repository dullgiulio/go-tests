## API

* ```/:table/fields``` All fields in table.

* ```/:table/query``` Query for data
  - JSON request:
  ```json
  {
	  tables: [ "table" ],
      fields: [ "a", "b", "c" ],
	  where: "a = ? AND b = ? OR c = ?",
	  values: { a: "blah", b: "okay", c: "test" },
	  groupBy: [ "b" ],
	  order: [ "c DESC" ],
  }
  ```
  - Returns response at text:
  ```
  $<bytes-after>\r\n
  {json-object-for-line}\r\n
  $<bytes-after>\r\n
  ```

## Queries

```sql
SELECT [source1:table.field] JOIN [source2:table] ON [source1:table.id] = [source2:table.id] WHERE [source2:table.field] = 100
```
