package query

const InsertEmail = "INSERT INTO vcommerce.emails(`email`, `created`) VALUES (?, now())"
const InsertUserID = "INSERT INTO vcommerce.userids(`user_id`, `created`) VALUES (?, now())"
const InsertUser = "INSERT INTO vcommerce.user(`unique_id`, `user_id`, `day_of_birth`, `profile_image`, `email`, `auth_type_json`, `meta_json`, `created`, `updated`) VALUES (?, ?, ?, ?, ?, ?, ?, now(), now())"
const InsertSession = "INSERT INTO vcommerce.session(`token`, `unique_id`, `created`, `updated`) VALUES (?, ?, now(), now())"
