{{$tableName := .TableName}}{{$now := .Now}}{{$idColumn := .IDColumn}}{{$titleColumn := .TitleColumn}}-- +goose Up
START TRANSACTION;
{{ range .Routes}}INSERT INTO `{{$tableName}}` (`{{$idColumn}}`,`{{$titleColumn}}`,`create_at`,`update_at`) VALUES ('{{.ID}}','{{.Title}}','{{$now}}','{{$now}}');
{{ end }}COMMIT;

-- +goose Down
START TRANSACTION;
{{ range .Routes}}DELETE FROM `{{$tableName}}` WHERE id = '{{.ID}}';
{{ end }}COMMIT;