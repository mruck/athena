# Athena
Athena is a prototype web application fuzzer.  It instruments the target application to collect metrics that inform parameter mutations, with the goal of detecting security violations.  Currently, Athena identifies SQL injection and unhandled Rails exceptions, flagging those that are potentially dangerous.

##### Benefits

- Code coverage metrics show what percentage of the code is safe, and what percentage is untested.  The metrics show the file and line number, with a goal in the future of being able to configure the fuzzer to cover those areas.
- On a security violation, the stack trace and request is provided so the bug can be triaged.  

##### Framework supported: 
Ruby on Rails

##### What is required from the user and why:
*User requirement*: A k8s pod spec with your containerized rails app.  Any other microservices should be in the pod spec or exposed on a port that the target can connect to.  
*Why*: Read world web applications are usually complex and talk to databases, etc.  A pod spec captures all this information and makes the environment reproducible.  We will patch the target image to point to our custom rails fork.  This rails fork contains collect information about the application and relays it to the fuzzer so that the fuzzer can intelligently mutate parameters.  The most important metrics collected by the rails fork are: 
1) Source code coverage: this indicates whether or not the fuzzer is making progress.  If the fuzzer is making progress, it will continue mutating parameters for the given endpoint.  If not, it will try a different endpoint.
2) Parameter accesses: this is not mandatory, but helps identify interesting parameters that the target is frequently accessing, as well as uninteresting parameters that the fuzzer shouldn't waste cycles mutating.
3) Database accesses: this helps Athena map parameters to tables and columns in the relational database, so that Athena can send parameters that stimulate the database and observer how user tainted behavior shows up in queries.  This allows Athena to detect sql injection.
4) Rails exceptions: Athena patches the target application so that all exceptions are logged and relayed back to the user.  Benign exceptions are whitelisted, while exceptions indicating a security problem are flagged.

*User requirement*: The target's backend must be postgres.  Currently, Athena only supports instrumenting postgres, but this can be extended in the future.
*Why*: This allows Athena to query the database and use values from the database to send in the parameters, which generally triggers more interesting behavior than randomly generated values.  Also as mentioned above, Athena can observe queries made by the target application to the database and look for malicious behavior.  

*User requirement*: Login script.
*Why*: The fuzzer can run unauthenticated, but only a small percentage of the application space will be explored.  More interesting behavior will be discovered if the fuzzer is logged in.

*User requirement*: HAR file.
*Why*: This is used as the “initial corpus”, it seeds the fuzzing engine with real human behavior.  This solves two problems: 1) realistic parameter values 2) route sequencing.  For example, if there were 2 routes, one to edit a post and one to create a post, the human will first hit the route to create a post then hit the route to edit the post.  The fuzzer won’t be able to do this ordering so having a sample set is very helpful.  In an ideal world, this corpus can be collected by proxying the QA team.

*User requirement*: Swagger spec.
*Why*: This tells the fuzzer the expected parameter types for each route, so its mutations are more human like.

##### Trophies
A security vulnerability was detected in Discourse and fixed in commits: [e2bcf5](https://github.com/discourse/discourse/commit/e2bcf55077be701a42f25651b26c4ac7028233c7),  [cac80cd](https://github.com/discourse/discourse/commit/cac80cdc3b5f847cfca6bf678e5a4c5e2837bbf3), [152238](https://github.com/discourse/discourse/commit/152238b4cff7ab4c4ce63ba26abd23b0abf05129).
Athena can also be used to find non-security 500s, however the market demand for that is ill defined.