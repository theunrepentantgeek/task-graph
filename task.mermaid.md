
:::mermaid
flowchart TD
  build["build"]
  build --> build_bin
  
  build_bin["build:bin"]
  build_bin --> build_dir
  
  build_dir["build:dir"]
  
  build_sbom["build:sbom"]
  build_sbom --> build_bin
  
  ci["ci"]
  ci --> test
  ci --> build_sbom
  
  default["default"]
  default --> build
  default --> test
  
  docs["docs"]
  docs --> docs_build
  
  docs_build["docs:build"]
  docs_build --> build
  
  gremlins["gremlins"]
  
  lint["lint"]
  
  samples["samples"]
  samples --> build
  
  test["test"]
  test --> unit-test
  test --> lint
  
  tidy["tidy"]
  tidy -.-> tidy_gofumpt
  tidy -.-> tidy_mod
  tidy -.-> tidy_lint
  
  tidy_gofumpt["tidy:gofumpt"]
  
  tidy_lint["tidy:lint"]
  
  tidy_mod["tidy:mod"]
  
  unit-test["unit-test"]
  unit-test --> build
  
  update-golden-files["update-golden-files"]
  
  classDef rule0 fill:lightblue
  class build_bin,build_dir,build_sbom rule0
  classDef rule1 fill:lightgreen
  class docs_build rule1
  classDef rule2 fill:lightyellow
  class tidy_gofumpt,tidy_lint,tidy_mod rule2
  classDef rule3 fill:lightsalmon
  class unit-test rule3
  classDef rule4 fill:lightpink
  class update-golden-files rule4
  classDef rule5 fill:lightgray
  class update-golden-files rule5
:::
