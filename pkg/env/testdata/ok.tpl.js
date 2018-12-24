function execute(tpl, params) {
  // Fetch image
  var fimg = tpl.FetchImage(params.fimage);
  var fcont = fimg.NewContainer(params.fcontainer);

  fcont.SetPorts(params.fport);
  fcont.SetLabel("ftest", params.flabel);

  // Build image
  var bimg = tpl.BuildImage(params.bimage);

  bimg.CopyDataToWorkspace("ws");
  bimg.InterpolateWorkspaceFile("ws", {
    "image": params.bimage,
    "cont": params.bcontainer
  });

  var bin = type.FromBase64("binary", params.binary);
  bimg.AddFileToWorkspace("binary", bin, 0755);

  var bcont = bimg.NewContainer(params.bcontainer);

  bcont.SetPorts(params.bport);
  bcont.SetLabel("btest", params.blabel);

  bcont.MountData("mount", "/mounted", {"interpolate": true});

  bcont.SetEnv("INTERPOLATE-ME",
               fmt("{{.ExternalAddress}}:{{.Self.ExposedPort %v}}",
                   params.bport));

  bcont.SetEnv("DONT-INTERPOLATE-ME", "WUT");

  bcont.AddReadinessCheck("net", {
    "protocol": "tcp",
    "address": fmt('{{.ExternalAddress}}:%v', params.rport)
  });
}
