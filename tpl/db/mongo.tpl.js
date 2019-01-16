function info() {
  return {
    "description": "MongoDB template",
    "parameters": {
      "init": {
        "description": "DB initialization queries",
        "type": "{\"db\": {\"collection\": [\"document\"]}}",
        "mandatory": false
      }
    }
  };
}

function execute(tpl, params) {
  var queries = [];

  for (var db in params.init) {
    var dbdata = params.init[db];

    queries.push(fmt('db = db.getSiblingDB("%s");', db));

    for (var col in dbdata) {
      var qs = dbdata[col];

      for (var i = 0; i < qs.length; i++) {
        queries.push(fmt("db.%s.insert(%s);", col, qs[i]));
      }
    }
  }

  var img = tpl.FetchImage("mongo:latest");
  var cont = img.NewContainer("mongo");
  var rules = queries.join("\n");

  cont.MountString(rules, "/docker-entrypoint-initdb.d/init.js", 0644, {});
  cont.SetLabel("mongo", "true");
  cont.SetPorts(27017);

  cont.AddReadinessCheck("net", {
    "protocol": "tcp",
    "address": '{{.ExternalAddress}}:{{.Self.ExposedPort 27017}}'
  });
}