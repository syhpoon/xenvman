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

  cont.MountData("domains.json", "/domains.json", {"interpolate": true});
  cont.SetPorts(8080);
  cont.SetLabel("xenv-discovery", "true");

  cont.SetCmd("/xenvman", "discovery",
              "--map=/domains.json",
              "--http-addr=:8080",
              "--dns-addr=:53",
              fmt("--recursors=%s", recursor)
  );
}