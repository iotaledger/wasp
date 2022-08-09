This folder contains a Docker-based setup to run your own Wasp development setup. 

# Starting
Run `docker-compose up` to start the setup. 

During the startup you might see a few failed restarts of Wasp with the message: 
`panic: error getting node event client: mqtt plugin not available on the current node`

This is normal, as Wasp starts faster than Hornet. Wasp retries the connection until it succeeds.

# Stopping
Press `Ctrl-C` to shut down the setup, but don't press it twice to force it. Otherwise, you can corrupt the Hornet database. 

You can also shut down the setup with `docker-compose down` in a new terminal. 

# Reset
Run `docker-compose down --remove-orphans` to shut down the nodes and to remove all databases.

The nodes will then be reachable under these ports:

- Wasp:
    - API: http://localhost:9090
    - Dashboard: http://localhost:7000 (username: wasp, password: wasp)
    - Nanomsg: tcp://localhost:5550 
    
- Hornet:
    - API: http://localhost:14266
    - External Peering: 15601/tcp
    - Dashboard: http://localhost:8082 (username: admin, password: admin)
    - Prometheus: http://localhost:9312/metrics