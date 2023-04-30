const json2ts = require("json-schema-to-typescript");
const fs = require("fs");

// compile from file
json2ts
  .compileFromFile("../../schemas/page.schema.json", { bannerComment: "" })
  .then((ts) => fs.writeFileSync("index.d.ts", ts));
