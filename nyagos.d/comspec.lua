for _,name in pairs{
    "assoc",
    "attrib",
    "dir",
    "mklink",
    "ren",
    "rename",
} do
    nyagos.alias[name] = "%COMSPEC% /c "..name.." $*"
end
