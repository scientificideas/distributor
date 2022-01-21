#### Distributor: a deterministic distribution of work between services

Distributor is a service that evenly distributes work items between services. 
If you have a certain amount of services, sharing responsibilities with each other, the Distributor helps you to avoid doing this manually. 
Moreover, the Distributor allows you to perform this spread deterministically: it provides the same division to the same list of services and work items regardless of other conditions (time, the instance of the Distributor itself, etc).

For example, you need to parse a certain amount of websites. 
You have a specific parser running on several machines. Each parser instance should parse its own subgroup of websites. 
There aren’t supposed to be any intersections: if a parser with ID “pX” is already parsing the website “sX”, none of the other parsers should perform this task again. 
Furthermore, each parser must receive the same subgroup of websites according to these requirements. If any of the parsers is down or the number of websites changes, the job will be redistributed.

<br>

#### Storage

The Distributor with the default storage implementation must be launched with Redis (a cluster or a single instance). 
Redis stores the list of services by a specific key (the list of such keys is comma-separated and indicated in **-services-namespaces=** or env **SERVICES_NAMESPACE**), the work units list (the key by which this list is stored is indicated in **-workunits-namespace=** or env **WORKUNITS_NAMESPACE**) and the so-called matching table (the key is set in **-distribution-namespace=** or in env **DISTRIBUTION_NAMESPACE**).

A matching table is a data structure, in which each service is matched with certain work units. For instance, the services list **[service1,service2,service3]** can be matched with the work units list **[workunit1,workunit2,workunit3,workunit4,workunit5,workunit6]** in the following way.

```
service1:workunit6,workunit4
service2:workunit3,workunit1
service3:workunit2,workunit5
```

The matching table is implemented using [Redis Hash](https://redis.io/topics/data-types).

<br>

#### The distribution of work between services

To complete the matching table, the Distributor collects the service list and the work units list from Redis and uses consistent hashing so that each service is deterministically matched to its work units. 
For that purpose, we enter work units into a hash ring (with the weight of 50 to distribute them evenly), sort the services list (so that the order in which services are entered into the list won’t affect the result), and then search for the closest corresponding value for each service. 
After detecting the closest corresponding value (a work unit related to this service), the work unit is deleted from the hash ring to avoid any possible repeated match (when two services are responsible for the same work unit).

After the initial distribution, you may need to perform a redistribution in case of service failure, so that the other services would take over the failed service’s work. 
To do so, the Distributor pings every service with a given interval (**-poll-interval=** or env **POLL_INTERVAL**) and considers every unsuccessful request as a denial, including timeout (which is set using **-ping-timeout=** or env **PING_TIMEOUT**). 
After detecting a denial, the work is redistributed according to the above algorithm, after which a new matching table is entered to Redis.

<br>

#### What is consistent hashing used for

An alternative to consistent hashing is the algorithm based on division with remainders:

**key ID mod len(buckets) = i**, where:
**key ID** is the hash/checksum of the service ID
**len(buckets)** is the number of work units
**i** is the index of a specific work unit in the array

In this case changes in the number of buckets (work units in our case) will lead to the redistribution of almost all keys. 
Consistent hashing allows us to avoid unnecessary redistributions (**n/m** of keys should be redistributed, where **n** is the number of keys and **m** is the number of buckets).

<br>

#### Instrumentation of services

Each service controlled by the Distributor must be available for pinging. For that to work, it’s necessary to implement the grpc-server to this service.

```
import "gitlab.n-t.io/atmz/distributor/pinger/grpc/server"
...
// grpc server for communicating with Distributor
go server.Inject(opts.conf.Host, opts.conf.Port)
```

Apart from that, at the start each service has to put itself on the list of services in Redis:

```
func (r *Redis) RegisterService(namespace, service string) error {
   // check it not registered already
   cmd := r.Client.LRange(namespace, 0, -1)
   var services []string
   if err := cmd.ScanSlice(&services); err != nil {
      return err
   }
   var isRegistered bool
   for _, serviceFound := range services {
      if serviceFound == service {
         isRegistered = true
      }
   }
   // add to list if not registered already
   if !isRegistered {
      return r.Client.LPush(namespace, service).Err()
   }
   return nil
}
```

After registering in Redis and launching the gRPC server, the service will be included in the matching table. Detecting the changed worklist for a service is the responsibility of the service itself. To do this, it needs to repeatedly check the Hash (hash table) in Redis.

```
func (r *Redis) GetMapField(key, field string) ([]string, error) {
   resp := r.Client.HGet(key, field)
   if resp.Err() != nil {
      return nil, resp.Err()
   }
   res, err := resp.Result()
   if err != nil {
      return nil, err
   }
   return strings.Split(res, ","), nil
}
```

Even though the necessity of checking the matching table by the service itself may seem excessive, it reduces the Distributor’s burden and eliminates an extra link in the chain of network requests (instead of using the service -> Distributor -> Redis pattern we only limit ourselves to service -> Redis). It also lowers the risk of inconsistent work distribution, caused by the rejection of Distributor instances, as well as it excludes the necessity of solving the task of detecting and choosing Distributor instances by the services.

Updating the work units list in Redis List is the responsibility of a separate independent service. 
It doesn’t have to communicate with other services, including Distributor. 
This service only has to add and update the Redis List elements under certain business logic, using the key indicated in **-workunits-namespace=** or env **WORKUNITS_NAMESPACE** when the Distributor is launched. 
This service cannot be written beforehand, neither can any default implementation be written. 
The reason is that it entirely depends on the distributed work units and their source. However, you don’t even have to create a separate service, it can be a simple module of your services, for which you create distributions. It’s all up to you. All you have to do is update the work units list in the storage.