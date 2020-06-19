# Wasp Publisher messages

Search for  "```publisher.Publish```" in the repo for exact places in the code. 

|Message|Format|
|:--- |:--- |
|SC bootup record has been saved in the registry | ```bootuprec <SC address> <SC color>``` |
|SC committee has been activated|```active_committee <SC address>```|
|SC committee dismissed|```dismissed_commitee <SC address>```|
|A new SC request reached the committee|```request_in <SC address> <request tx ID> <request block index>```|
|SC request has been processed (i.e. corresponding state update was confirmed)|```request_out <SC address> <request tx ID> <request block index> <state index> <seq number in the batch> <batch size>```|
|State transition (new state committed to DB)| ```state <SC address> <state index> <batch size> <state tx ID> <state hash> <timestamp>```|
|VM (processor) initialized succesfully|```vmready <SC address> <program hash>```|

