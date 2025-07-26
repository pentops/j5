Null and Empty Values
=====================

In J5, fields can be Required, Explicitly Optional or Implicitly Optional.

In the J5 schema file, Explicitly Optional has the `?` marker Required has the `!` marker.

An untagged field is Implicitly Optional, meaning that 'null' (or not
specified), and empty values are all equivalent.

A field which is Explicitly Optional can be set to 'null', or the empty value
(0, "", false), these are considered different values. These recieve the
`optional` tag in proto schemas, resulting in pointers in go structs.

A required field must have a non-default value, meaning "", 0 and false are not valid,
nor is null or undefined.

A very special case is a field which is both Required and Explicitly Optional
(?!), meaning that "", 0 and false are all acceptable values, but must be
explicitly set in the input JSON, a null or undefined value is invalid. These do
not recieve the `optional` tag in proto schemas, and are not pointers in go
structs, this only effects validation and encoding.


# JSON Encoding

In JSON, a string can be `null`, `undefined`, or an empty string.
All of these are falsy/empty, but they can be considered different values.

Undefined means it's not populated at all 

```
{}
```

Null is explicitly set to null

```json
{
  "key": null
}
```

An empty string is a string with no characters

```json
{
  "key": ""
}
```
## Include Empty

A setting in the codec defines treatment of these values when encoding.

### Default Behavior

Fields with non-default values will always be encoded.

- Required fields will always be encoded, even when the value is "", 0 or false.
- Explicitly Optional fields which are null/undefined will not be encoded
- Explicitly Optional fields with "", 0, false will be encoded
- Implicitly Optional fields which are null/undefined will not be encoded
- Implicitly Optional fields with "", 0, false will not be encoded


| Field Type          | Null | Zero    |
|---------------------|------|---------|
| Required            | N/A  | Yes     |
| Explicitly Optional | No   | Yes     |
| Implicitly Optional | No   | No      |

### Include Empty = True

When the codec setting `includeEmpty` is set to true, the behavior changes so
that all keys will always be encoded as something, be that `null` `""` or `0`.

| Field Type          | Null        | Zero    |
|---------------------|-------------|---------|
| Required            | N/A         | Yes     |
| Explicitly Optional | As null     | Yes     |
| Implicitly Optional | As Default  | Yes     |

Timestamp and Date fields have no zero value, no assumption of 1970.
Alongisde Objects and Oneofs, null is the zero value, so implicitly and
explicitly optional carries no distinction.

Enums treat the zero value and the undefined value equally, and from there they
work like strings, but the zero value is encoded as "UNSPECIFIED" rather than
"". Either is accepted.


