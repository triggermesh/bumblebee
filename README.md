### Transformation prototype

Transformation is an addressable CR based on Knative Serving and 
[kazaam](https://github.com/qntfy/kazaam) package. When you create 
Transformation object controller creates Knative Service that 
accepts CloudEvents, applies 
[transformation](https://github.com/qntfy/kazaam#specification-support) 
and responses back with new CloudEvent.

Below you can find a sample manifest which has PingSource that 
generates a json:

```
{
  "foo": "bar", 
  "extra": "data", 
  "project": "Triggermesh"
}
```

The CE with JSON payload being routed to Transformation CR where 
it gets modified according to the Specs and then being routed to 
event display service. If you look into evet display logs you'll
see new CE payload:

```
{
  "not-foo": "bar",
  "sub": {
    "project": "hello,null",
    "uuid": "62ab680f-66ee-47e1-95dd-cab19462c2ee"
  }
}
```

Transformation Spec format is something that will be changed
after we figure out our requirements. Also, kazaam package 
showed some inconsistent behavior so it either needs to be
better tested and documented or replaced with its analog.

How CE being routed:

```
+------------------+        +-------------------+      +------------------+
| Source           |        |                   |      |                  |
|------------------|        |                   +----->|                  |
|                  +------->| Broker            |      | Transformation   |
| Pingsource       |        |                   |<-----+                  |
|                  |        |                   |      |                  |
+------------------+        +--------+----------+      +------------------+
                                     |
                                     |
                                     v
                            +-------------------+
                            | Target            |
                            |-------------------|
                            |                   |
                            | Event-display     |
                            |                   |
                            +-------------------+
```

Full list of objects:

```
apiVersion: eventing.knative.dev/v1beta1
kind: Broker
metadata:
  annotations:
    eventing.knative.dev/broker.class: MTChannelBasedBroker
  name: transformation-demo
spec:
  config:
    apiVersion: v1
    kind: ConfigMap
    name: config-br-default-channel
    namespace: knative-eventing

---

apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: event-display
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-releases/knative.dev/eventing-contrib/cmd/event_display

---

apiVersion: eventing.knative.dev/v1beta1
kind: Trigger
metadata:
  name: trigger-transformation
spec:
  broker: transformation-demo
  filter:
    attributes:
      type: dev.knative.sources.ping
  subscriber:
    ref:
      apiVersion: flow.triggermesh.io/v1alpha1
      kind: Transformation
      name: demo
      namespace: default

---

apiVersion: eventing.knative.dev/v1beta1
kind: Trigger
metadata:
  name: trigger-eventdisplay
spec:
  broker: transformation-demo
  filter: 
    attributes:
      type: ce.after.transformation
  subscriber:
    ref:
      apiVersion: serving.knative.dev/v1
      kind: Service
      name: event-display
      namespace: default

---

apiVersion: sources.knative.dev/v1alpha2
kind: PingSource
metadata:
  name: ping-source
spec:
  schedule: "*/1 * * * *"
  jsonData: '{
    "foo": "bar", 
    "extra": "data", 
    "project": "Triggermesh"
    }'
  sink:
    ref:
      apiVersion: eventing.knative.dev/v1beta1
      kind: Broker
      name: transformation-demo

---

apiVersion: flow.triggermesh.io/v1alpha1
kind: Transformation
metadata:
  name: demo
spec:
  events:
  - transform: '[
      {"operation": "delete", "spec": {"paths": ["extra"]}},
      {"operation": "shift", "spec": {"not-foo": "foo"}},
      {"operation": "concat", "spec": {"sources": [{"value": "hello"}, {"path": "project"}], "targetPath": "sub.project", "delim": ","}},
      {"operation": "uuid", "spec": {"sub.uuid": {"version": 4}}}
      ]'
    targetType: ce.after.transformation

```