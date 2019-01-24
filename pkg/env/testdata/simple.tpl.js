function execute(tpl, params) {
  // Fetch image
  var img = tpl.FetchImage(params.image);
  var cont = img.NewContainer(params.container);

  if(params.mount === true) {
    cont.MountData("mount", "/mounted", { "interpolate": true });
  } else {
    cont.SetLabel("simple", "true");
  }
}
