package query

const InsertSession = "INSERT INTO vcommerce.session(`token`, `unique_id`, `created`, `updated`) VALUES (?, ?, now(), now())"

const InsertEmail = "INSERT INTO vcommerce.emails(`email`, `created`) VALUES (?, now())"
const InsertUserID = "INSERT INTO vcommerce.userids(`user_id`, `created`) VALUES (?, now())"
const InsertUser = "INSERT INTO vcommerce.user(`unique_id`, `user_id`, `day_of_birth`, `cell_phone_number`, `profile_image`, `email`, `meta_json`, `created`, `updated`) VALUES (?, ?, ?, ?, ?, ?, ?, now(), now())"

const InsertSellerAuth = "INSERT INTO vcommerce.seller(`unique_id`, `seller_type`, `company_registration_number`, `owner_name`, `company_name`, `channel_name`, `channel_url`, `channel_description`, `bank_name`, `bank_account_number`, `uploaded_file_path`, `created`, `updated`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now(), now())"
const InsertSellerChannel = "INSERT INTO vcommerce.seller_channel(`channel_name`, `created`) VALUES (?, now())"
const InsertSellerRegistration = "INSERT INTO vcommerce.seller_registration(`unique_id`, `authentication`, `created`, `updated`) VALUES (?, ?, now(), now())"

const InsertVideoList = "INSERT INTO vcommerce.video_info(`video_id`, `video_url`, `serve_ready`, `created`, `updated`) VALUES (?, ?, 0, now(), now())"
const InsertProductCategoryInfo = "INSERT INTO vcommerce.product_category(`product_id`, `category_json`, `created`, `json`) VALUES (?, ?, now(), now())"
const InsertProductSale = "INSERT INTO vcommerce.product(`product_id`, `unique_id`, `video_list_json`, `title`, `base_price`, `base_amount`, `option_json`, `deleted`, `created`) VALUES (?, ?, ?, ?, ?, ?, ?, 0, now())"

const InsertCart = "INSERT INTO vcommerce.cart(`cart_id`, `unique_id`, `product_id`, `selected_json`, `created`) VALUES (?, ?, ?, ?, now())"
const DeleteCart = "DELETE FROM vcommerce.cart WHERE cart_id=? AND unique_id=?"

const InsertReview = "INSERT INTO vcommerce.review(`review_id`, `product_id`, `unique_id`, `thumb_up_down_id`, `body`, `media_info_json`, `star`, `created`, `updated`) VALUES(?, ?, ?, ?, ?, ?, ?, now(), now())"
