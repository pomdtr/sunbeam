#!/usr/bin/env -S deno run -A

import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

async function search(query: string): Promise<sunbeam.ListItem[]> {
  if (query.length === 0) return [];

  const response = await fetch(
    "https://search.nixos.org/backend/latest-42-nixos-23.05/_search",
    {
      body: JSON.stringify({
        from: 0,
        size: 50,
        sort: [
          {
            _score: "desc",
            package_attr_name: "desc",
            package_pversion: "desc",
          },
        ],
        aggs: {
          package_attr_set: { terms: { field: "package_attr_set", size: 20 } },
          package_license_set: {
            terms: { field: "package_license_set", size: 20 },
          },
          package_maintainers_set: {
            terms: { field: "package_maintainers_set", size: 20 },
          },
          package_platforms: {
            terms: { field: "package_platforms", size: 20 },
          },
          all: {
            global: {},
            aggregations: {
              package_attr_set: {
                terms: { field: "package_attr_set", size: 20 },
              },
              package_license_set: {
                terms: { field: "package_license_set", size: 20 },
              },
              package_maintainers_set: {
                terms: { field: "package_maintainers_set", size: 20 },
              },
              package_platforms: {
                terms: { field: "package_platforms", size: 20 },
              },
            },
          },
        },
        query: {
          bool: {
            filter: [
              {
                term: { type: { value: "package", _name: "filter_packages" } },
              },
              {
                bool: {
                  must: [
                    { bool: { should: [] } },
                    { bool: { should: [] } },
                    { bool: { should: [] } },
                    { bool: { should: [] } },
                  ],
                },
              },
            ],
            must: [
              {
                dis_max: {
                  tie_breaker: 0.7,
                  queries: [
                    {
                      multi_match: {
                        type: "cross_fields",
                        query,
                        analyzer: "whitespace",
                        auto_generate_synonyms_phrase_query: false,
                        operator: "and",
                        _name: "multi_match_test",
                        fields: [
                          "package_attr_name^9",
                          "package_attr_name.*^5.3999999999999995",
                          "package_programs^9",
                          "package_programs.*^5.3999999999999995",
                          "package_pname^6",
                          "package_pname.*^3.5999999999999996",
                          "package_description^1.3",
                          "package_description.*^0.78",
                          "package_longDescription^1",
                          "package_longDescription.*^0.6",
                          "flake_name^0.5",
                          "flake_name.*^0.3",
                        ],
                      },
                    },
                    {
                      wildcard: {
                        package_attr_name: {
                          value: "*test*",
                          case_insensitive: true,
                        },
                      },
                    },
                  ],
                },
              },
            ],
          },
        },
      }),
      method: "POST",
      headers: {
        'Content-Type': 'application/json',
        'User-Agent': 'Sunbeam',
        Authorization: "Basic YVdWU0FMWHBadjpYOGdQSG56TDUyd0ZFZWt1eHNmUTljU2g="
      },
    }
  );

  if (!response.ok) throw new Error("Failed to search: " + response.statusText);

  const json = await response.json();
  return json.hits.hits.map((hit: any) => ({
    title: hit._source.package_attr_name,
    subtitle: hit._source.package_description || "",
    accessories: [hit._source.package_pversion],
    actions: [
      {
        "type": "open",
        "title": "Open in NixOS Search",
        "url": `https://search.nixos.org/packages?query=${encodeURIComponent(query)}&show=${encodeURIComponent(hit._source.package_attr_name)}`
      },
      {
        "type": "open",
        "title": "Open Homepage",
        "url": hit._source.package_homepage[0],
      },
      {
        type: "copy",
        title: "Copy Package Name",
        key: "c",
        exit: true,
        text: hit._source.package_attr_name,
      }
    ]
  }));
}

if (Deno.args.length === 0) {
  const manifest: sunbeam.Manifest = {
    title: "Nixpkgs Search",
    root: ["search"],
    commands: [
      {
        name: "search",
        title: "Search Packages",
        mode: "search",
      },
    ],
  };
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const payload = JSON.parse(Deno.args[0]) as sunbeam.Payload;
if (payload.command !== "search") {
  console.error(`Unknown command: ${payload.command}`);
  Deno.exit(0);
}

if (!payload.query) {
  const list: sunbeam.List = { emptyText: "Search for a package" };
  console.log(JSON.stringify(list));
  Deno.exit(0);
}

const items = await search(payload.query);
const list: sunbeam.List = {
  items,
};
console.log(JSON.stringify(list));
