# Print the values the sentry module handed to the Sentry SDK in this Odoo process.
#
# Usage, inside the container:
#   odoo-bin shell -c /home/odoo/.odoorc --no-http < sentry_values.py
#
# No database is needed: the sentry module initializes the SDK from its post_load hook, which
# runs when the server-wide modules are loaded, before any database is opened.
import sentry_sdk

client = sentry_sdk.get_client()
print("sentry active: %s" % client.is_active())
for key in ("dsn", "environment", "release", "dist"):
    print("%-12s= %r" % (key, client.options.get(key) if client.is_active() else None))
