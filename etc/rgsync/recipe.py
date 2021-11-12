from rgsync import RGWriteBehind, RGWriteThrough
from rgsync.Connectors import MySqlConnector, MySqlConnection

'''
Create MySQL connection object
'''
connection = MySqlConnection('username', 'password', 'mysql:3306/database')

'''
Create MySQL object_instance connector
'''
objectInstancesConnector = MySqlConnector(connection, 'object_instances', 'id')

objectInstancesMappings = {
	'parent_id': 'parent_id',
	'inside_object_instance_id': 'inside_object_instance_id',
	'name': 'name',
	'short_description': 'short_description',
	'long_description': 'long_description',
	'description': 'description',
	'flags': 'flags',
	'wear_location': 'wear_location',
	'item_type': 'item_type',
	'value_1': 'value_1',
	'value_2': 'value_2',
	'value_3': 'value_3',
	'value_4': 'value_4'
}

RGWriteBehind(GB, keysPrefix='object', mappings=objectInstancesMappings, connector=objectInstancesConnector, name='ObjectInstancesWriteBehind', version='99.99.99')