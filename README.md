# CloudEvents Transformation

Transformation is an addressable CR based on Knative Serving 
aimed on flexible CloudEvents modifications. When you create 
Transformation object controller creates Knative Service that 
accepts CloudEvents, applies transformation and replies with 
new CloudEvent or forwards it to another addressable resource.

Current Transformation engine support following basic operations

## Operations

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

##### Store

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

## Events routing

The CE with JSON payload being routed to Transformation CR where
it gets modified according to the Specs and then being routed back
to sender:

![bumblebee](https://user-images.githubusercontent.com/13515865/94548224-35f97400-0272-11eb-9d22-7dcd1ce0d639.png)


## Example

[Sample](config/samples) directory contains two examples, let's walk through one named "multi-target".
All commands below are idempotent and assume that your current path is the repository root directory.

First of all, you need to create the Knative Eventing Broker:

```
kubectl apply -f config/samples/broker.yaml
```

Next, open [config/samples/multi-target/githubsource.yaml](config/samples/multi-target/githubsource.yaml) and set `accessToken` and `secretToken` values as described in the [documentation](https://knative.dev/docs/eventing/samples/github-source/#create-github-tokens), change `ownerAndRepository` to your Github username and repository you want to track. Create the resources:

```
kubectl apply -f config/samples/multi-target/githubsource.yaml
```

After the Github source is created, open [config/samples/multi-target/googlesheet-target.yaml](config/samples/multi-target/googlesheet-target.yaml), paste the [credentials JSON key](https://github.com/triggermesh/knative-targets/blob/master/docs/googlesheet.md#prerequisites) into the `googlesheet` Secret's `credentials` field, scroll down and update GoogleSheetTarget `id` value as described in the [readme](https://github.com/triggermesh/knative-targets/blob/master/docs/googlesheet.md#creating-a-googlesheet-target). Create the resources:

```
kubectl apply -f config/samples/multi-target/googlesheet-target.yaml
```

Now let's edit our second target - [config/samples/multi-target/slack-target.yaml](config/samples/multi-target/slack-target.yaml). The `slacktarget` secret needs to have a Slack token which you can obtain by following [this](https://github.com/triggermesh/knative-targets/blob/master/docs/slack.md#creating-the-slack-app-bot-and-token-secret) document. Also, you should specify which Slack channel should receive our messages by setting its name in the `slack-transformation` object, line 56. After it's done, create the resources:

```
kubectl apply -f config/samples/multi-target/slack-target.yaml
```

The final step is to create the Githubsource events transformations:

```
kubectl apply -f config/samples/multi-target/github-transformation.yaml
```

Here is what we essentially created:

![example1](https://user-images.githubusercontent.com/13515865/94557645-a909e700-0280-11eb-868b-6592b6bc8d9c.png)

**(1)** - Gihubsource [githubsource-transformation-demo](config/samples/multi-target/githubsource.yaml) receives the [issues](https://docs.github.com/en/free-pro-team@latest/developers/webhooks-and-events/webhook-events-and-payloads#webhook-payload-example-when-someone-edits-an-issue) and the [push](https://docs.github.com/en/free-pro-team@latest/developers/webhooks-and-events/webhook-events-and-payloads#webhook-payload-example-33) event webhooks from the Github and sends it to the `transformation-demo` Broker.

**(2)**, **(3)** - [first](config/samples/multi-target/github-transformation.yaml) "layer" of Triggers with Transformations are throwing away all unwanted data, standardizing different Github events into a single format, e.g.:

```
{
  "object": "tzununbekov",
  "subject": "triggermesh/bumblebee",
  "verb": "created issue \"Transformation tests are failing\"",
}
```

Original CloudEvent type changed to a common `io.triggermesh.transformation.github` value.

**(4)**, **(5)** - target-specific Triggers are picking up serialized Github Events and wrapping them into the payloads digestable by the final targets, e.g. [SlackTarget](config/samples/multi-target/slack-target.yaml):

```
{
  "channel": "github-events-channel",
  "text": "tzununbekov at triggermesh/bumblebee: created issue \"Transformation tests are failing\"",
}
```
Types are set to the values to match the corresponding Target Trigger only.

**(6)**, **(7)** - finally, Events are passing through the filters of the Target Triggers and being delivered to its destinations - GoogleSheet table and Slack channel in our case.

At first, this approach may seem a bit cumbersome, but taking into account a number of possible Github [Events](https://docs.github.com/en/free-pro-team@latest/developers/webhooks-and-events/webhook-events-and-payloads) multiplied by the number of possible additional Targets, decoupling producer from consumer starts to make sense.

## Support

We would love your feedback and help on this project, so don't hesitate to let us know what is wrong and how we could improve them, just file an [issue](https://github.com/triggermesh/bumblebee/issues/new) or join those of us who are maintaining them and submit a [PR](https://github.com/triggermesh/bumblebee/compare)

## Commercial Support

TriggerMesh Inc. supports this project commercially, email info@triggermesh.com to get more details.

## Code of Conduct

This plugin is by no means part of [CNCF](https://www.cncf.io/) but we abide by its [code of conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md)
