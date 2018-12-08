function execute(tpl, params) {
  var img = tpl.FetchImage(params.image);
  var cont = img.NewContainer(params.container);

  cont.SetLabel("test", params.label)
}
