## Transformation prototype

Transformation is an addressable CR based on Knative Serving 
aimed on flexible CloudEvents modifications. When you create 
Transformation object controller creates Knative Service that 
accepts CloudEvents, applies transformation and replies with 
new CloudEvent.

Current Transformation engine support following basic operations

### Operations

#### Delete

Delete CE keys or objects.

Example 1.

Remove a key.

```
spec:
  data:
  - name: delete
    paths:
    - key: foo
    - key: array[1].foo
    - key: foo.array[5]
```

Example 2.

Remove a "foo" key only if its value is equal to "bar". 

```
spec:
  data:
  - name: delete
    paths:
    - key: foo
      value: bar
```

Example 3.

Recursively remove all keys with specified value.

```
spec
  data:
  - name: delete
    paths:
    - value: leaked password
```

#### Add

Add new or override existing CE keys.

Example 1.

Override Cloud Event type. This operation can be used to implement
complex Transformation logic with multiple Triggers and CE type
filtering.

```
spec:
  context:
  - name: add
    paths: 
    - key: type
      value: ce.after.transformation
```

Example 2.

Create a new object with nested structure. Value "42" will be 
converted to integer.

```
spec:
  data:
  - name: add
    paths:
    - key: The.Ultimate.Questions.Answer
      value: "42"
```

Example 3.

Create arrays or modify existing ones. "True" value will be
converted to boolean and added as a second item of a new array
"array" in a new object "newObject". "1337" will be added as
an integer with a new key "newKey" as a first item of and
existing array "commits".

```
spec:
  data:
  - name: add
    paths:
    - key: newObject.array[2]
      value: "true"
    - key: commits[1].newKey
      value: "1337"
```

#### Shift

Move existing CE values to new keys.

Example 1.

Move value from "foo" key to "bar"

```
spec:
  data:
  - name: shift
    paths:
    - key: foo:bar
```

Example 2.

Move key only if its value is equal to "bar".

```
spec:
  data:
  - name: shift
    paths:
    - key: old:new
      value: bar
```

Example 3.

Shift supports nested objects and arrays:

```
spec:
  data:
  - name: shift
    paths:
    - key: array[0].id:newArray[1].newId
    - key: object.list[0]:newItem
```

#### Store

Store CE value as a Pipeline variable. Useful in combination with 
other operations. Variables are shared across pipelines and in 
theory may be used as a key and/or as a value.

Example.

Store CE type and source and add them into headers array in a payload.
Also set a new CE type and save the original one in context extensions.

```
spec:
  context:
  - name: store
    paths:
    - key: $ceType
      value: type
    - key: $ceSource
      value: source
  - name: add
    paths:
    - key: type
      value: ce.after.transformation
    - key: extensions.OriginalType
      value: $ceType
  data:
  - name: add
    paths:
    - key: headers[0].source
      value: $ceSource
    - key: headers[1].type
      value: $ceType
```

### Events routing

The CE with JSON payload being routed to Transformation CR where 
it gets modified according to the Specs and then being routed back
to sender:


```
+------------------+        +-------------------+      +------------------+
| Source           |        |                   |      |Transformation1   |
+------------------+        |                   +----->-------------------+
|                  +------->+ Broker            |      | If CE.type is FOO|
| Pingsource       |        |                   +<-----+ Set KEY1 = VAL1  |
|                  |        |                   |      | Set CE.type = BAR|
+------------------+        +-------------------+      +------------------+
                            If CE.type is READY
                            Send it to a target        +------------------+
                                      +                |Transformation2   |
                            +---------v---------+      +------------------+
                            | Target            |      | If CE.type is BAR|
                            +-------------------+      | Set KEY2 = VAL2  |
                            |                   |      | Set CE.type = BAZ|
                            | Event+display     |      +------------------+
                            |                   |
                            +-------------------+       ...
                                                       +------------------+
                                                       |TransformationN   |
                                                       +------------------+
                                                       |...               |
                                                       |                  |
                                                       |                  |
                                                       +------------------+
```

### Sample

[Sample](config/samples) directory contains manifests to deploy
full set of objects including Broker, Event-display, Triggers and
[Transformation](config/samples/transformation.yaml) to see how
it works.
