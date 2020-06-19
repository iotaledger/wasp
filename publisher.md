# Wasp Publisher messages

Wasp publishes important events via Nanomsg message stream (just like ZMQ is used in IRI. Possibly  in the future ZMQ and MQTT publishers will be supported too).

Anyone can subscribe to the Nanomsg output stream of the node. In Golang you can use `packages/subscribe` package provided in Wasp for this.
The publisher's output port can be configured in ```config.json``` like this:
```
  "nanomsg":{
    "port": 5550
  } 
```

Search for  "```publisher.Publish```" in the repo for exact places in the code where messages are published. 

Currently supported messages and formats (space separated list of strings):

|Message|Format|
|:--- |:--- |
|SC bootup record has been saved in the registry | ```bootuprec <SC address> <SC color>``` |
|SC committee has been activated|```active_committee <SC address>```|
|SC committee dismissed|```dismissed_commitee <SC address>```|
|A new SC request reached the node|```request_in <SC address> <request tx ID> <request block index>```|
|SC request has been processed (i.e. corresponding state update was confirmed)|```request_out <SC address> <request tx ID> <request block index> <state index> <seq number in the batch> <batch size>```|
|State transition (new state committed to DB)| ```state <SC address> <state index> <batch size> <state tx ID> <state hash> <timestamp>```|
|VM (processor) initialized succesfully|```vmready <SC address> <program hash>```|

