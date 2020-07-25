# Let's talk Probes

So I was wondering why I see so many cases where people create Deployments for Kubernetes and took the effort of 
defining the readinessProbe and livenessProbe. This is a good thing, since Kubernetes should be able to heal application 
once pods don't respond the way they should. 

However, what I do see, is that in most cases, the same endpoint is used for both the readiness as the liveness Probes.
And that, is a bit redundant. So after a talk on the Kubernetes Slack, I noticed that there is some confusion as to what 
the difference is between, and how they should be used.

As such, I decided to make a simple demo to show what the exact difference is between the two, and at the same time show 
a new Probe as well, named startupProbe, which was added as an Alpha feature in 1.16, and graduated to beta in 1.18.

## A simple Go web server

So for this example I created a simple web server which has 4 different endpoints.

Namely:
* /startupProbe
* /livenessProbe
* /readinessProbe
* /maxReadinessCountProbe
* /startJob

The functions name should be a clear enough indication as to what their functions are, apart from `/startjob` which I will 
elaborate on further up.

The Probe endpoints will start with giving up 503's for a given duration. The extend of 
that duration can be determined by setting the appropriate ENV values, namely: `WAIT_STARTUP_TIME`, `WAIT_READINESS_TIME` and `WAIT_LIVENESS_TIME`.
After that duration has passed, they will give off 200's.

Then we also have an added feature to the `/readinessProbe` endpoint. Once a GET is done against `/startjob` the `/readinessProbe` endpoint 
will start to give off 503 for a given duration. The duration can be set by adjusting the  `JOB_DURATION_TIME` ENV variable, 
after the duration has passed, it will also start giving off 200's again.

`IS_READINESS_EQUALS_LIVENESS` is a setting which makes Readiness time interval equal to the Liveness interval.

`MAX_READINESS_COUNT` overrides the `WAIT_READINESS_TIME` duration and once the number of readiness probes reaches the max 
count, the `/maxReadinessCountProbe` endpoint would start replying with 503.


The default values are:  
`ENV WAIT_STARTUP_TIME 30`  
`ENV WAIT_LIVENESS_TIME 60`  
`ENV WAIT_READINESS_TIME 90`  
`ENV JOB_DURATION_TIME 20`  
`ENV IS_READINESS_EQUALS_LIVENESS false`  
`ENV MAX_READINESS_COUNT 2`


## Building the container

The Dockerfile is included in the repo. Building it, is simple. Simply run:  
`docker build . -t health-probes`

If you want your custom docker image to be copied and deployed on other target systems. Run:
`docker save health-probes:latest | gzip > health-probes_latest.tar.gz`
to get the latest docker image as a tar.gz file. Followed by scp-ing the tar.gz file over to intended system.


To load the copied tar.gz file as a docker image on the target system, run the following on the
target system:
`docker load < health-probes_latest.tar.gz`

To run the latest health-probe image :  
`docker run -P -p 8080:8080 health-probes`


Then you can access it using localhost:
* http://localhost:8080/readinessProbe
* http://localhost:8080/livenessProbe
* http://localhost:8080/startupProbe
* http://localhost:8080/maxReadinessCountProbe
* http://localhost:8080/startJob
