// Available params:
// init  :: init - Database initialization queries
//
// Types:
// init :: {"<db>": {"<collection>": ["<document>"]}}

function execute(tpl, params) {
  var queries = [];

  for(var db in params.init) {
    var dbdata = params.init[db];

    queries.push(fmt('db = db.getSiblingDB("%s");', db));

    for(var col in dbdata) {
      var qs = dbdata[col];

      for(var i=0; i < qs.length; i++) {
        queries.push(fmt("db.%s.insert(%s);", col, qs[i]));
      }
    }
  }

  var img = tpl.FetchImage(fmt("mongo:latest"));
  var cont = img.NewContainer("mongo");
  var rules = queries.join("\n");

  cont.MountString(rules, "/docker-entrypoint-initdb.d/init.js", 0644, {});
  cont.SetLabel("mongo", "true");
}