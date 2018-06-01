tt = read.table("jumps.txt", header=T)

png("range-count-log.png")
plot(tt$JumpRange, tt$Count, log="xy", pch=20, main="Jump Count (log)", xlab="Jump Range (LY)", ylab="Jump Count")
dev.off()

png("range-count.png")
plot(tt$JumpRange, tt$Count, xlim=c(10,40), ylim=c(0, 200), pch=20, main="Jump Count", xlab="Jump Range (LY)", ylab="Jump Count")
dev.off()

png("range-efficiency.png")
plot(tt$JumpRange, tt$Efficiency, pch=20, ylim=c(0,1), main="Efficiency", xlab="Jump Range (LY)", ylab="Efficiency")
dev.off()

png("range-efficiency-zoom.png")
plot(tt$JumpRange, tt$Efficiency, pch=4, xlim=c(8,20), ylim=c(0,1), main="Efficiency", xlab="Jump Range (LY)", ylab="Efficiency")
dev.off()
