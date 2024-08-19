-- Подключение и создание БД, если не существует
box.cfg({ listen = 3301 })
box.schema.user.create("storage", { password = "admin", if_not_exists = true })
box.schema.user.grant("storage", "super", nil, nil, { if_not_exists = true })

-- Создание таблицы хранилища
local kv_storage = box.schema.space.create("kv_storage", { if_not_exists = true })
kv_storage:format({
	{ name = "key", type = "string" },
	{ name = "value", type = "any" },
})
kv_storage:create_index("primary", {
	if_not_exists = true,
	parts = { "key" },
})

-- Создание таблицы пользователей
local kv_users = box.schema.space.create("kv_users", { if_not_exists = true })
kv_users:format({
	{ name = "username", type = "string" },
	{ name = "password", type = "string" },
})
kv_users:create_index("primary", {
	if_not_exists = true,
	parts = { "username" },
})

-- Создаем пользователя, если не существует
if #kv_users:select({ "admin" }) == 0 then
	kv_users:insert({ "admin", "presale" })
end
