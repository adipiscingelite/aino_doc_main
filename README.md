# backend

Base URL = http://localhost:8080

# authentication
- /login -> (user_email, password) (required all)
            Member -> middleware role_code = 'M'
            Admin -> middleware role_code = 'A'
            Superadmin -> middleware role_code = 'SA'


# add user
//add division, role, and application first
- /division/add
- /application/add
- /role/add
- /user/add -> (application_uuid, division_uuid, role_uuid required) 