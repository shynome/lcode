import yaml from "https://esm.sh/yaml@2.2.1";

const f = ".github/workflows/go-ossf-slsa3-publish.yml";
const root = await Deno.readTextFile(f).then((r) => yaml.parse(r));

const matrix = root.jobs.build.strategy.matrix;
/**@type {string} */
const fileFomat = root.jobs.build.with["config-file"];

const tpl = await Deno.readTextFile(".slsa-goreleaser.yml");

let tasks = [];
for (let os of matrix.os) {
  for (let arch of matrix.arch) {
    let n = tpl.replace("{{goos}}", os).replace("{{goarch}}", arch);
    let f = fileFomat
      .replace("${{matrix.os}}", os)
      .replace("${{matrix.arch}}", arch);
    let task = Deno.writeTextFile(f, n);
    tasks.push(task);
  }
}

await Promise.all(tasks);
console.log("generator finish");
