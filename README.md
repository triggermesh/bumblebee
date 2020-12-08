# Bumblebee: A CloudEvents Transformer

Bumblebee allows you to transform [CloudEvents](https://cloudevents.io/) via an object of kind `Transformation`. A Transformation object creates an addressable custom resource based on Knative Serving 
aimed at flexible CloudEvents modifications. When you create a
Transformation object, the controller creates a Knative Service that 
accepts CloudEvents, applies the declared transformation and replies with a
new CloudEvent or forwards it to another addressable resource.

## Installation

Bumblebee can be compiled and deployed from source with [ko](https://github.com/google/ko):

```
ko apply -f ./config
```

You can verify that it installed successfully by checking the CRD:

```
$ kubectl get crd transformations.flow.triggermesh.io
transformations.flow.triggermesh.io                  2020-08-19T13:13:09Z
```

And checking that the controller is running:

```
$ kubectl get pods -n transformation -l app=transformation-controller
transformation-controller-6bdc658bf8-pwblp                1/1     Running   0          5d19h
```

A custom resource of kind `Transformation` can now be created, check a [sample](https://github.com/triggermesh/bumblebee/blob/master/config/samples/simple-event/transformation.yaml).

## Specification

Bumblebee's API specification consists of three parts: optional Sink reference and two transformation sections called "context" and "data" for corresponding [CloudEvents](https://github.com/cloudevents/spec/blob/v1.0/spec.md) components. If a Bumblebee object (i.e `Transformation`) has a sink then the resulting events are forwarded to the referenced object, otherwise, they will be sent back to the event producer. "context" and "data" transformation operations are applied on the event in the order they are listed in the spec with one exception: "store". The "store" operation runs before the rest to be able to collect variables for the runtime. 

## Operations

Currently Bumblebee supports the following basic transformation operations:

### Delete

Delete CE keys or objects.

##### Example 1

Remove a key.

```yaml
spec:
  data:
  - operation: delete
    paths:
    - key: foo
    - key: array[1].foo
    - key: foo.array[5]
```

##### Example 2

Remove a "foo" key only if its value is equal to "bar". 

```yaml
spec:
  data:
  - operation: delete
    paths:
    - key: foo
      value: bar
```

##### Example 3

Recursively remove all keys with specified value.

```yaml
spec:
  data:
  - operation: delete
    paths:
    - value: leaked password
```

##### Example 4

Delete everything. Useful for composing completely new CE
using stored variables.

```yaml
spec:
  data:
  - operation: delete
    paths:
    - key:
```

### Add

Add new or override existing CE keys.

##### Example 1

Override Cloud Event type. This operation can be used to implement
complex Transformation logic with multiple Triggers and CE type
filtering.

```yaml
spec:
  context:
  - operation: add
    paths:
    - key: type
      value: ce.after.transformation
```

##### Example 2

Create a new object with nested structure.

```yaml
spec:
  data:
  - operation: add
    paths:
    - key: The.Ultimate.Questions.Answer
      value: "42"
```

##### Example 3

Create arrays or modify existing ones. "True" will be added as 
a second item of a new array "array" in a new object "newObject".
"1337" will be added as a new key "newKey" as a first item of an
existing array "commits".

```yaml
spec:
  data:
  - operation: add
    paths:
    - key: newObject.array[2]
      value: "true"
    - key: commits[1].newKey
      value: "1337"
```

##### Example 4

"Add" operation supports value composing from variables and
static strings.

```yaml
spec:
  data:
  - operation: add
    paths:
    - key: id
      value: ce-$source-$id
```

### Shift

Move existing CE values to new keys.

##### Example 1

Move value from "foo" key to "bar"

```yaml
spec:
  data:
  - operation: shift
    paths:
    - key: foo:bar
```

##### Example 2

Move key only if its value is equal to "bar".

```yaml
spec:
  data:
  - operation: shift
    paths:
    - key: old:new
      value: bar
```

##### Example 3

Shift supports nested objects and arrays:

```yaml
spec:
  data:
  - operation: shift
    paths:
    - key: array[0].id:newArray[1].newId
    - key: object.list[0]:newItem
```

### Store

Store CE value as a pipeline variable. Useful in combination with
the other operations. The variables are shared between the "context"
and the "data" parts of the transformation pipeline.

##### Example 1

Store CE type and source and add them into headers array in a payload.
Also set a new CE type and save the original one in context extensions.

```yaml
spec:
  context:
  - operation: store
    paths:
    - key: $ceType
      value: type
    - key: $ceSource
      value: source
  - operation: add
    paths:
    - key: type
      value: ce.after.transformation
    - key: extensions.OriginalType
      value: $ceType
  data:
  - operation: add
    paths:
    - key: headers[0].source
      value: $ceSource
    - key: headers[1].type
      value: $ceType
```

## Sample with Event Routing

Transformations are useful to modify the payload and CloudEvent context attributes when an event is routed to a Target (aka event sink) that needs to receive a specific event type and payload. The CloudEvent can be routed to a Transformation addressable via a specific Trigger where
it gets modified according to the declared transformation and then gets routed to its final destination via a second Trigger as depicted in the figure below:

![bumblebee](https://user-images.githubusercontent.com/13515865/94548224-35f97400-0272-11eb-9d22-7dcd1ce0d639.png)

The [Sample](config/samples) directory contains examples.

## Support

We would love your feedback and help on this project, so don't hesitate to let us know what is wrong and how we could improve them, just file an [issue](https://github.com/triggermesh/bumblebee/issues/new) or join those of us who are maintaining them and submit a [PR](https://github.com/triggermesh/bumblebee/compare)

## Commercial Support

TriggerMesh Inc. supports this project commercially, email info@triggermesh.com to get more details.

## Code of Conduct

This plugin is by no means part of [CNCF](https://www.cncf.io/) but we abide by its [code of conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md)
