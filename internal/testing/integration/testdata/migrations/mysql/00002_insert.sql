-- +goose Up
-- +goose StatementBegin
INSERT INTO owners (owner_name, owner_type)
VALUES
  ('lucas', 'user'),
  ('space', 'organization');
-- +goose StatementEnd

INSERT INTO owners (owner_name, owner_type)
VALUES
  ('james', 'user'),
  ('TheAmirhosssein', 'organization');

INSERT INTO repos (repo_full_name, repo_owner_id)
VALUES
  ('james/rover', (SELECT owner_id FROM owners WHERE owner_name = 'james')),
  ('TheAmirhosssein/goose', (SELECT owner_id FROM owners WHERE owner_name = 'TheAmirhosssein'));

-- +goose Down
-- +goose StatementBegin
DELETE FROM owners;
-- +goose StatementEnd
