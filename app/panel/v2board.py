from Panel.interface import PanelInterface
import mysql.connector

class V2Board(PanelInterface):


    """
    update ip in v2board, via sql
    """

    TABLE_NAME = ["v2_server_v2ray", "v2_server_trojan", "v2_server_shadowsocks"]

    def __init__(
        self, 
        db_host:str,
        db_name:str,
        db_user:str,
        db_password:str,
    ):

        self.db_host = db_host
        self.db_name = db_name
        self.db_user = db_user
        self.db_password = db_password
    
        # test connect
        conn = self.__connect_to_db()
        self.__disconnect(conn)

    def get_node_address_list(self):
        db = self.__connect_to_db()
        cur = db.cursor()
        addr_list = [ ]

        for table_name in self.TABLE_NAME:
            sql = f"""
                SELECT host, port FROM {table_name}
                """
            cur.execute(sql)
            rows = cur.fetchall()
            for row in rows:
                addr_list.append(f"{row[0]}:{row[1]}")
            
        db.commit()
        self.__disconnect(db)
        return addr_list

    def update_node_address(self, old_addr, new_addr):

        db = self.__connect_to_db()
        cur = db.cursor()

        def _(oldip, newip, table_name, entry_name):
            sql = f"""
                UPDATE {table_name}
                SET {entry_name} = "{newip}"
                WHERE {entry_name} = "{oldip}"
                """
            cur.execute(sql)

        for table_name in self.TABLE_NAME:
            _(old_addr, new_addr, table_name, "host")

        db.commit()
        self.__disconnect(db)

    
    def __connect_to_db(self):
        return mysql.connector.connect(
            host=self.db_host,
            user=self.db_user,
            password=self.db_password,
            db=self.db_host
        )

    def __disconnect(self, connection):
        connection.disconnect()