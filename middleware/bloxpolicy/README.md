# bloxpolicy

`bloxpolicy` enables request filtering based on the blox policy engine.

## Syntax

~~~
bloxpolicy
~~~

* With no arguments, the blox policy engine is assumed to be available at
  the URL: http://localhost:10000

~~~
bloxpolicy endpoint
~~~

* endpoint is the URL to use as the blox policy engine API.


## Examples

Filter all requests using the blox policy engine at url.

~~~
bloxpolicy http://localhost:9000
~~~
