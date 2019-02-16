function execute(tpl, params) {
  type.EnsureNumber("number", params.number);

  var img = tpl.FetchImage("import-image");
  var cont = img.NewContainer("import2-cont");
}
