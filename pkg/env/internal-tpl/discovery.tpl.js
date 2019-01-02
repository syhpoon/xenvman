// Service discovery template
//
// Parameters:
// recursor :: string {"1.1.1.1"} - External DNS server to use

function execute(tpl, params) {
  type.EnsureString("recursor", params.dnsRecursor);

  var img = tpl.FetchImage("syhpoon/xenvman:latest");
  var cont = img.NewContainer("discovery");

  var recursor = "1.1.1.1";

  if (type.IsDefined(params.recursor)) {
    recursor = params.recursor;
  }

  var port = 8080;

  cont.MountData("domains.json", "/domains.json", {"interpolate": true});
  cont.SetPorts(port);
  cont.SetLabel("xenv-discovery", "true");
  cont.SetLabel("xenv-discovery-port", fmt("%d", port));

  cont.SetCmd("/xenvman", "discovery",
              "--map=/domains.json",
              fmt("--http-addr=:%d", port),
              "--dns-addr=:53",
              fmt("--recursors=%s", recursor)
  );

  cont.AddReadinessCheck("http", {
    "url": fmt('http://{{.ExternalAddress}}:{{.Self.ExposedPort %d}}/health', port),
    "codes": [200]
  });
}