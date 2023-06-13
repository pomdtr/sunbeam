const json2ts = require("json-schema-to-typescript");
const fs = require("fs");

const wrapper = `
export function List(list: List) {
  return list;
}

export function Detail(detail: Detail) {
  return detail;
}

export function Item(item: Listitem) {
  return item;
}

export function Action(action: Action) {
  return action;
}

export function Input(input: Input) {
  return input;
}
`;

// compile from file
json2ts
  .compileFromFile("../../schemas/page.schema.json", {
    bannerComment: "",
  })
  .then((types) => [types, wrapper].join("\n\n"))
  .then((ts) => fs.writeFileSync("index.ts", ts));
