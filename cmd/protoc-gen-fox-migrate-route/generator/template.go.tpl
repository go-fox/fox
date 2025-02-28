{{$tableName := .TableName}}
package {{.PackageName}}

import (
    "ariga.io/atlas/sql/migrate"
    "context"
    "entgo.io/ent/dialect"
    "entgo.io/ent/dialect/sql/schema"
    {{.Imports}}
)

// {{.FuncName}} 编写插入
func {{.FuncName}}(dir *migrate.LocalDir) error {
    w := &schema.DirWriter{Dir: dir}
    client := ent.NewClient(ent.Driver(schema.NewWriteDriver(dialect.MySQL, w)))
    if err := client.{{$tableName}}.CreateBulk(
        {{ range .Routes}}client.{{$tableName}}.Create().SetID("{{.ID}}").SetTitle("{{.Title}}"),
        {{ end }}
    ).Exec(context.Background()); err != nil {
        return err
    }
    return w.FlushChange("insert_route", "Inset the route")
}