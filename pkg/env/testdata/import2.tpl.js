function execute(tpl, params) {
  type.EnsureNumber("number", params.number);
  type.EnsureListOfNumbers("numbers1", params.numbers1);
  type.EnsureListOfNumbers("numbers2", params.numbers2);
  type.EnsureListOfNumbers("numbers3", params.numbers3);

  var img = tpl.FetchImage("import-image");
  var cont = img.NewContainer("import2-cont");
}
