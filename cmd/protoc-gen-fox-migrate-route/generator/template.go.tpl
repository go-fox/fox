{{$tableName := .TableName}}{{$now := .Now}}{{$idColumn := .IDColumn}}{{$titleColumn := .TitleColumn}}-- +goose Up
-- +goose StatementBegin
START TRANSACTION;
{{ range .Routes}}INSERT INTO `{{$tableName}}` (`{{$idColumn}}`,`{{$titleColumn}}`,`create_at`,`update_at`) VALUES ('{{.ID}}','{{.Title}}','{{$now}}','{{$now}}');
{{ end }}COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
START TRANSACTION;
{{ range .Routes}}DELETE FROM `{{$tableName}}` WHERE id = '{{.ID}}';
{{ end }}COMMIT;
-- +goose StatementEnd