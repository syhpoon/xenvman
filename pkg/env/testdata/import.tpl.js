function execute(tpl, params) {
  var img = tpl.FetchImage("import-image");
  var cont = img.NewContainer("import-cont");

  import_template("import2", {
    "number": params.number,
    "numbers1": [123],
    "numbers2": [3.14],
    "numbers3": [123, 3.14],
  });
  import_template("import3", {});
}
