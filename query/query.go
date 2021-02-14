package query

const InsertEmail = "INSERT INTO vcommerce.emails(`email`, `created`) VALUES (?, now())"
const InsertUserID = "INSERT INTO vcommerce.userids(`user_id`, `created`) VALUES (?, now())"
const InsertUser = "INSERT INTO vcommerce.user(`unique_id`, `user_id`, `day_of_birth`, `profile_image`, `email`, `auth_type_json`, `meta_json`, `created`, `updated`) VALUES (?, ?, ?, ?, ?, ?, ?, now(), now())"
const InsertSession = "INSERT INTO vcommerce.session(`token`, `unique_id`, `created`, `updated`) VALUES (?, ?, now(), now())"
const InsertSellerAuth = "INSERT INTO vcommerce.seller(`unique_id`, `seller_type`, `company_registration_number`, `owner_name`, `company_name`, `channel_name`, `channel_url`, `channel_description`, `bank_name`, `bank_account_number`, `created`, `updated`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now(), now())"
const InsertSellerChannel = "INSERT INTO vcommerce.seller_channel(`channel_name`, `created`) VALUES (?, now())"
const InsertSellerRegistration = "INSERT INTO vcommerce.seller_registration(`unique_id`, `authentication`, `created`, `updated`) VALUES (?, ?, now(), now())"
