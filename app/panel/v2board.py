import mysql.connector


class V2board:

    """
    update ip in v2board, via sql
    """

    TABLE_NAME = ["v2_server_v2ray", "v2_server_trojan", "v2_server_shadowsocks"]

    ENTRY_NAME = ["host"]

    def __init__(self, config):

        self.config = config["panel"]["v2board"]

    def connect_to_db(self):
        return mysql.connector.connect(
            host=self.config["host"],
            user=self.config["user"],
            password=self.config["password"], 
            db=self.config["db"]
        )

    def disconnect(self, connection):
        connection.disconnect()


    def update_ip(self, oldip, newip):

        db = self.connect_to_db()
        cur = db.cursor()

        def _(oldip, newip, table_name, entry_name):
            sql = f"""
                UPDATE {table_name} 
                SET {entry_name} = "{newip}" 
                WHERE {entry_name} = "{oldip}"
                """
            cur.execute(sql)

        for table_name in self.TABLE_NAME:
            for entry_name in self.ENTRY_NAME:
                _(oldip, newip, table_name, entry_name)

        db.commit()
        self.disconnect(db)
