function info() {
  return {
    "description": "CockroachDB template",
    "parameters": {
      "init": {
        "description": "DB initialization queries",
        "type": "string[]",
        "mandatory": false,
      }
    }
  };
}

function execute(tpl, params) {
  var img = tpl.FetchImage("cockroachdb/cockroach:v2.1.3");
  var cont = img.NewContainer("crdb");

  cont.SetLabel("cockroachdb", "true");
  cont.MountData("init.sh", "/init.sh", {});
  cont.SetEntrypoint("/init.sh");
  cont.SetPorts(8080, 26257);

  if(type.IsDefined(params.init)) {
    cont.MountString(params.init.join("\n"), "/init.sql", 0644, {});
  }

  cont.AddReadinessCheck("http", {
    "url": 'http://{{.ExternalAddress}}:{{.SelfExposedPort 8080}}/health',
    "codes": [200],
    "retry_limit": 10,
  });
}