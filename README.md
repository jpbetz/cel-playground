CEL Playground
--------------

Evaluate a CEL expression against a yaml file:

```sh
./cel-playground eval --expr "deployment.spec.replicas > 10" --variables deployment=test/deployment.yaml
```

Run a playground server:

```sh
./cel-playground serve
```

Send a request to the server:

```sh
curl localhost:8080/eval \
  -H "Content-Type: application/yaml" \
  -X POST \
  -d '{"expression": "1 < x", "variables": {"x": 2}}'
```

TODO
----

- [ ] Include a Web UI in the playground server
- [ ] Share playgrounds with links (encode data in link for small playgrounds, need storage backend for larger playgrouns)
- [ ] Support OpenAPIv3 schemas and compilation (e.g. `--schemas deployment=test/deployment-openapiv3.yaml)
- [ ] Support context cancelation and cost limits