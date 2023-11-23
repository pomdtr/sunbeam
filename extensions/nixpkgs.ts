#!/usr/bin/env -S deno run -A

import * as sunbeam from "https://deno.land/x/sunbeam/mod.ts";

async function search(searchText: string): Promise<SearchResult[]> {
  if (searchText.length === 0) return [];

  const [searchSize, branchName] = [20, "unstable"]

  const queryFields = [
    "package_attr_name^9",
    "package_attr_name.edge^9",
    "package_pname^6",
    "package_pname.edge^6",
    "package_attr_name_query^4",
    "package_attr_name_query.edge^4",
    "package_description^1.3",
    "package_description.edge^1.3",
    "package_longDescription^1",
    "package_longDescription.edge^1",
    "flake_name^0.5",
    "flake_name.edge^0.5",
    "package_attr_name_reverse^7.2",
    "package_attr_name_reverse.edge^7.2",
    "package_pname_reverse^4.800000000000001",
    "package_pname_reverse.edge^4.800000000000001",
    "package_attr_name_query_reverse^3.2",
    "package_attr_name_query_reverse.edge^3.2",
    "package_description_reverse^1.04",
    "package_description_reverse.edge^1.04",
    "package_longDescription_reverse^0.8",
    "package_longDescription_reverse.edge^0.8",
    "flake_name_reverse^0.4",
    "flake_name_reverse.edge^0.4",
  ];

  const reversedSearchText = [...searchText].reverse().join("");
  const query = {
    size: Math.trunc(searchSize),
    sort: [{ _score: "desc" }, { package_attr_name: "desc" }, { package_pversion: "desc" }],
    query: {
      bool: {
        filter: [{ term: { type: { value: "package", _name: "filter_packages" } } }],
        must: [
          {
            dis_max: {
              tie_breaker: 0.7,
              queries: [
                {
                  multi_match: {
                    type: "cross_fields",
                    query: searchText,
                    analyzer: "whitespace",
                    auto_generate_synonyms_phrase_query: false,
                    operator: "and",
                    _name: `multi_match_${searchText}`,
                    fields: queryFields,
                  },
                },
                {
                  multi_match: {
                    type: "cross_fields",
                    query: reversedSearchText,
                    analyzer: "whitespace",
                    auto_generate_synonyms_phrase_query: false,
                    operator: "and",
                    _name: `multi_match_${reversedSearchText}`,
                    fields: queryFields,
                  },
                },
                { wildcard: { package_attr_name: { value: `*${searchText}*` } } },
              ],
            },
          },
        ],
      },
    },
  };

  const response = await fetch(
    `https://nixos-search-7-1733963800.us-east-1.bonsaisearch.net/latest-42-nixos-${branchName}/_search`,
    {
      method: "post",
      headers: {
        Authorization: "Basic YVdWU0FMWHBadjpYOGdQSG56TDUyd0ZFZWt1eHNmUTljU2g=",
        "Content-Type": "application/json",
      },
      body: JSON.stringify(query),
    }
  );

  const json = (await response.json()) as
    | {
      hits: {
        hits: {
          _id: string;
          _source: {
            package_pname: string;
            package_attr_name: string;
            package_attr_set: string;
            package_outputs: string[];
            package_default_output: string | null;
            package_description: string | null;
            package_homepage: string[];
            package_pversion: string;
            package_platforms: string[];
            package_position: string;
            package_license: { fullName: string; url: string | null }[];
          };
        }[];
      };
    }
    | { error: { reason: string }; status: number }
    | { code: string; message: string };

  if ("code" in json) {
    throw new Error(json.message);
  } else if ("error" in json) {
    throw new Error(json.error.reason);
  } else if (!response.ok) {
    throw new Error(response.statusText);
  }

  return json.hits.hits.map(({ _source: result, _id: id }) => {
    return {
      id,
      name: result.package_pname,
      attrName: result.package_attr_name,
      description: result.package_description,
      version: result.package_pversion,
      homepage: result.package_homepage,
      source:
        result.package_position &&
        `https://github.com/NixOS/nixpkgs/blob/nixos-unstable/${result.package_position.replace(/:([0-9]+)$/, "")}`,
      outputs: result.package_outputs,
      defaultOutput: result.package_default_output,
      platforms: result.package_platforms.filter((platform) =>
        ["x86_64-linux", "aarch64-linux", "i686-linux", "x86_64-darwin", "aarch64-darwin"].includes(platform)
      ),
      licenses: result.package_license.map((license) => {
        let url: string | null;
        try {
          url = license.url ?? new URL(license.fullName).href;
        } catch {
          url = null;
        }
        return { name: license.fullName, url };
      }),
    };
  });
}

interface SearchResult {
  id: string;
  name: string;
  attrName: string;
  description: string | null;
  version: string;
  homepage: string[];
  source: string | null;
  outputs: string[];
  defaultOutput: string | null;
  platforms: string[];
  licenses: { name: string; url: string | null }[];
}

if (Deno.args.length === 0) {
  const manifest: sunbeam.Manifest = {
    title: "Nixpkgs Search",
    root: ["search"],
    commands: [
      {
        name: "search",
        title: "Search",
        mode: "search"
      }
    ]
  }
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const payload = JSON.parse(Deno.args[0]) as sunbeam.Payload;
if (payload.command == "search") {
  if (payload.query) {
    const results = await search(payload.query);
    const list: sunbeam.List = {
      items: results.map((result) => (
        {
          title: result.name,
          subtitle: result.description || "",
          accessories: [result.version],
          actions: [
            {
              type: "copy",
              title: "Copy Name",
              text: result.name,
              exit: true
            },
            {
              type: "open",
              title: "Open in Browser",
              target: `https://search.nixos.org/packages?channel=unstable&show=${encodeURIComponent(result.name)}`,
              exit: true
            }
          ]
        }
      ))
    }
    console.log(JSON.stringify(list));
  } else {
    const list: sunbeam.List = { emptyText: "Search for a package" }
    console.log(JSON.stringify(list));
  }
}


