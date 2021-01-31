package query

const InsertUserQuery = "INSERT INTO vcommerce.user(`uuid`, `nickname`, `day_of_birth`, `profile_image`, `email`, `auth_type_json`, `meta_json`,`registered`, `created`, `updated`) VALUES (?, ?, ?, ?, ?, ?, ?, 1, now(), now())"
