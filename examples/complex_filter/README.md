# Complex Filter Example

This example demonstrates how to use the `Joiner` and `Wherer` interface under the same `Filter` instance to create a
complex filter.

## Database Schema

### People Table

| id | name | age |
|----|------|-----|
| 1  | Ben  | 25  |
| 2  | John | 30  |

### Contacts Table

| id | person_id | email               |
|----|-----------|---------------------|
| 1  | 1         | ben@example.com     |
| 2  | 2         | john@exampletwo.com |
