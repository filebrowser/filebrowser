SELECT city, COUNT(id) AS users_count
FROM users
WHERE group_name = 'salesman'
AND created > '2011-05-21'
GROUP BY 1
ORDER BY 2 DESC