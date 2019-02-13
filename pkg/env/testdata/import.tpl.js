function execute(tpl, params) {
  var img = tpl.FetchImage("import-image");
  var cont = img.NewContainer("import-cont");

  import_template("import2", {});
  import_template("import3", {});
}
