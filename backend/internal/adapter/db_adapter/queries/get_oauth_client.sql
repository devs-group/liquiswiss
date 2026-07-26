SELECT client_id, client_name, redirect_uris
FROM oauth_clients
WHERE client_id = ?;
